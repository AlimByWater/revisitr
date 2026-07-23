package wallet

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math/big"
	"net/http"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"revisitr/internal/entity"
)

// ── Apple PassKit pass.json structures ──────────────────────────────────────

type applePass struct {
	FormatVersion       int              `json:"formatVersion"`
	PassTypeIdentifier  string           `json:"passTypeIdentifier"`
	SerialNumber        string           `json:"serialNumber"`
	TeamIdentifier      string           `json:"teamIdentifier"`
	OrganizationName    string           `json:"organizationName"`
	Description         string           `json:"description"`
	ForegroundColor     string           `json:"foregroundColor,omitempty"`
	BackgroundColor     string           `json:"backgroundColor,omitempty"`
	LabelColor          string           `json:"labelColor,omitempty"`
	LogoText            string           `json:"logoText,omitempty"`
	Barcode             *appleBarcode    `json:"barcode,omitempty"`
	StoreCard           appleStoreCard   `json:"storeCard"`
	AuthenticationToken string           `json:"authenticationToken"`
	WebServiceURL       string           `json:"webServiceURL,omitempty"`
}

type appleBarcode struct {
	Format          string `json:"format"`
	Message         string `json:"message"`
	MessageEncoding string `json:"messageEncoding"`
}

type appleStoreCard struct {
	PrimaryFields    []appleField `json:"primaryFields"`
	SecondaryFields  []appleField `json:"secondaryFields,omitempty"`
	AuxiliaryFields  []appleField `json:"auxiliaryFields,omitempty"`
	BackFields       []appleField `json:"backFields,omitempty"`
}

type appleField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value any    `json:"value"`
}

// ── PKCS7 ASN.1 structures ─────────────────────────────────────────────────

var (
	oidData          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidSignedData    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidSHA1          = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
	oidRSAEncryption = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidContentType   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	oidSigningTime   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
)

type pkcs7ContentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,optional,tag:0"`
}

type pkcs7IssuerAndSerial struct {
	IssuerName   asn1.RawValue
	SerialNumber *big.Int
}

type pkcs7SignerInfo struct {
	Version                   int
	IssuerAndSerialNumber     pkcs7IssuerAndSerial
	DigestAlgorithm           pkcs7AlgorithmID
	AuthenticatedAttributes   []pkcs7Attribute `asn1:"set,implicit,tag:0"`
	DigestEncryptionAlgorithm pkcs7AlgorithmID
	EncryptedDigest           []byte
}

type pkcs7AlgorithmID struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

type pkcs7Attribute struct {
	Type   asn1.ObjectIdentifier
	Values []asn1.RawValue `asn1:"set"`
}

type pkcs7SignedData struct {
	Version          int
	DigestAlgorithms []pkcs7AlgorithmID `asn1:"set"`
	ContentInfo      pkcs7ContentInfo
	Certificates     asn1.RawValue `asn1:"optional,implicit,tag:0"`
	SignerInfos      []pkcs7SignerInfo `asn1:"set"`
}

// ── PassGenerator ───────────────────────────────────────────────────────────

type PassGenerator struct{}

func NewPassGenerator() *PassGenerator {
	return &PassGenerator{}
}

// GeneratePass builds a .pkpass bundle for Apple Wallet.
func (g *PassGenerator) GeneratePass(
	pass *entity.WalletPass,
	cfg *entity.WalletConfig,
	clientQRCode string,
	orgName string,
	webServiceURL string,
) ([]byte, error) {
	// 1. Decode Apple certificate from credentials
	certData, err := base64.StdEncoding.DecodeString(cfg.Credentials["certificate"])
	if err != nil {
		return nil, fmt.Errorf("decode certificate: %w", err)
	}

	password := ""
	if pwd, ok := cfg.Credentials["password"]; ok {
		password = pwd
	}

	privateKey, certificate, caCerts, err := pkcs12.DecodeChain(certData, password)
	if err != nil {
		return nil, fmt.Errorf("pkcs12 decode: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("certificate private key is not RSA")
	}

	// Ensure the Apple WWDR intermediate is in the signature chain — .p12 bundles
	// frequently ship only the leaf, and Apple rejects passes without it.
	if wwdr, err := appleWWDRCAG4(); err == nil {
		present := false
		for _, c := range caCerts {
			if c.Equal(wwdr) {
				present = true
				break
			}
		}
		if !present {
			caCerts = append(caCerts, wwdr)
		}
	}

	// 2. Build pass.json
	passJSON, err := g.buildPassJSON(pass, cfg, clientQRCode, orgName, webServiceURL)
	if err != nil {
		return nil, fmt.Errorf("build pass.json: %w", err)
	}

	// 3. Build manifest (SHA-1 hashes)
	manifest := map[string]string{}
	manifest["pass.json"] = sha1Hex(passJSON)

	// Add images — fetch logo once so the manifest hash and the zipped bytes match.
	var logoData []byte
	if cfg.Design.LogoURL != "" {
		if data, err := fetchImage(cfg.Design.LogoURL); err == nil {
			logoData = data
			manifest["logo.png"] = sha1Hex(logoData)
		}
	}

	// Generate icon
	iconData, err := g.generateIcon(cfg.Design.BackgroundColor)
	if err != nil {
		return nil, fmt.Errorf("generate icon: %w", err)
	}
	manifest["icon.png"] = sha1Hex(iconData)

	// Generate QR code
	var qrData []byte
	if clientQRCode != "" {
		qrData, err = g.generateQRCode(clientQRCode)
		if err != nil {
			return nil, fmt.Errorf("generate qr: %w", err)
		}
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	// 4. Create PKCS7 signature
	allCerts := append([]*x509.Certificate{certificate}, caCerts...)
	signature, err := g.signManifest(manifestJSON, rsaKey, allCerts)
	if err != nil {
		return nil, fmt.Errorf("sign manifest: %w", err)
	}

	// 5. Zip everything
	return g.zipPass(passJSON, manifestJSON, iconData, qrData, logoData, signature)
}

func (g *PassGenerator) buildPassJSON(
	pass *entity.WalletPass,
	cfg *entity.WalletConfig,
	clientQRCode string,
	orgName string,
	webServiceURL string,
) ([]byte, error) {
	if orgName == "" {
		orgName = cfg.Design.OrganizationName
	}

	ap := applePass{
		FormatVersion:      1,
		PassTypeIdentifier: cfg.Credentials["pass_type_id"],
		SerialNumber:       pass.SerialNumber,
		TeamIdentifier:     cfg.Credentials["team_id"],
		OrganizationName:   orgName,
		Description:        cfg.Design.Description,
		ForegroundColor:    rgbColor(cfg.Design.ForegroundColor),
		BackgroundColor:    rgbColor(cfg.Design.BackgroundColor),
		LabelColor:         rgbColor(cfg.Design.LabelColor),
		AuthenticationToken: pass.AuthToken,
		WebServiceURL:      webServiceURL,
		StoreCard: appleStoreCard{
			PrimaryFields: []appleField{
				{Key: "balance", Label: "БАЛАНС", Value: pass.LastBalance},
			},
			SecondaryFields: []appleField{
				{Key: "level", Label: "УРОВЕНЬ", Value: pass.LastLevel},
			},
			BackFields: []appleField{
				{Key: "client", Label: "КЛИЕНТ", Value: fmt.Sprintf("#%d", pass.ClientID)},
			},
		},
	}

	if clientQRCode != "" {
		ap.Barcode = &appleBarcode{
			Format:          "PKBarcodeFormatQR",
			Message:         clientQRCode,
			MessageEncoding: "iso-8859-1",
		}
	}

	return json.Marshal(ap)
}

func (g *PassGenerator) signManifest(manifestJSON []byte, key *rsa.PrivateKey, certs []*x509.Certificate) ([]byte, error) {
	// Compute SHA-1 hash of manifest
	hash := sha1.Sum(manifestJSON)

	// Build authenticated attributes
	now := time.Now().UTC()

	contentTypeVal, err := asn1.Marshal(oidData)
	if err != nil {
		return nil, fmt.Errorf("marshal contentType value: %w", err)
	}

	digestVal, err := asn1.Marshal(hash[:])
	if err != nil {
		return nil, fmt.Errorf("marshal digest value: %w", err)
	}

	timeVal, err := asn1.Marshal(now)
	if err != nil {
		return nil, fmt.Errorf("marshal time value: %w", err)
	}

	attrs := []pkcs7Attribute{
		{Type: oidContentType, Values: []asn1.RawValue{{FullBytes: contentTypeVal}}},
		{Type: oidMessageDigest, Values: []asn1.RawValue{{FullBytes: digestVal}}},
		{Type: oidSigningTime, Values: []asn1.RawValue{{FullBytes: timeVal}}},
	}

	// DER-encode authenticated attributes for signing. Per RFC 5652 §5.4 the
	// signature is computed over the attributes with an EXPLICIT SET OF tag
	// (0x31), not the IMPLICIT [0] tag they carry inside SignerInfo.
	wrapped, err := asn1.Marshal(struct {
		Attrs []pkcs7Attribute `asn1:"set"`
	}{Attrs: attrs})
	if err != nil {
		return nil, fmt.Errorf("marshal auth attrs: %w", err)
	}
	var aaRaw asn1.RawValue
	if _, err := asn1.Unmarshal(wrapped, &aaRaw); err != nil {
		return nil, fmt.Errorf("unwrap auth attrs: %w", err)
	}
	aaSetDER := aaRaw.Bytes // the SET OF encoding (0x31 ...)

	// RSA-SHA1 sign the SHA-1 digest of the DER-encoded attributes.
	aaHash := sha1.Sum(aaSetDER)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, aaHash[:])
	if err != nil {
		return nil, fmt.Errorf("rsa sign: %w", err)
	}

	// Build the certificates [0] IMPLICIT field: the concatenated certificate
	// DER encodings, tagged as context-specific [0] (not a SEQUENCE/SET).
	var certBytes []byte
	for _, c := range certs {
		certBytes = append(certBytes, c.Raw...)
	}

	issuerName := asn1.RawValue{FullBytes: certs[0].RawIssuer}

	// Build SignerInfo
	signerInfo := pkcs7SignerInfo{
		Version: 1,
		IssuerAndSerialNumber: pkcs7IssuerAndSerial{
			IssuerName:   issuerName,
			SerialNumber: certs[0].SerialNumber,
		},
		DigestAlgorithm: pkcs7AlgorithmID{
			Algorithm:  oidSHA1,
			Parameters: asn1.RawValue{Tag: asn1.TagNull},
		},
		AuthenticatedAttributes: attrs,
		DigestEncryptionAlgorithm: pkcs7AlgorithmID{
			Algorithm:  oidRSAEncryption,
			Parameters: asn1.RawValue{Tag: asn1.TagNull},
		},
		EncryptedDigest: sig,
	}

	// Build SignedData
	signedData := pkcs7SignedData{
		Version: 1,
		DigestAlgorithms: []pkcs7AlgorithmID{
			{Algorithm: oidSHA1, Parameters: asn1.RawValue{Tag: asn1.TagNull}},
		},
		ContentInfo: pkcs7ContentInfo{
			ContentType: oidData,
		},
		Certificates: asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: certBytes},
		SignerInfos:  []pkcs7SignerInfo{signerInfo},
	}

	sdDER, err := asn1.Marshal(signedData)
	if err != nil {
		return nil, fmt.Errorf("marshal signedData: %w", err)
	}

	// Build top-level ContentInfo. The SignedData is carried in a [0] EXPLICIT
	// wrapper, so use Bytes (not FullBytes) to let the marshaler emit the tag.
	ci := pkcs7ContentInfo{
		ContentType: oidSignedData,
		Content:     asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true, Bytes: sdDER},
	}

	return asn1.Marshal(ci)
}

func (g *PassGenerator) zipPass(passJSON, manifestJSON, iconData, qrData, logoData, signature []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	files := []struct {
		name string
		data []byte
	}{
		{"pass.json", passJSON},
		{"manifest.json", manifestJSON},
		{"signature", signature},
		{"icon.png", iconData},
	}
	if len(logoData) > 0 {
		files = append(files, struct {
			name string
			data []byte
		}{"logo.png", logoData})
	}

	for _, file := range files {
		f, err := w.Create(file.name)
		if err != nil {
			return nil, fmt.Errorf("zip %s: %w", file.name, err)
		}
		if _, err := f.Write(file.data); err != nil {
			return nil, fmt.Errorf("write %s: %w", file.name, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}

	return buf.Bytes(), nil
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func (g *PassGenerator) generateIcon(bgColor string) ([]byte, error) {
	c := parseHexColor(bgColor, color.RGBA{80, 80, 80, 255})
	img := image.NewRGBA(image.Rect(0, 0, 58, 58))
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode icon: %w", err)
	}
	return buf.Bytes(), nil
}

func (g *PassGenerator) generateQRCode(data string) ([]byte, error) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("new qr: %w", err)
	}
	qr.DisableBorder = true
	pngData, err := qr.PNG(400)
	if err != nil {
		return nil, fmt.Errorf("qr png: %w", err)
	}
	return pngData, nil
}

func sha1Hex(data []byte) string {
	h := sha1.Sum(data)
	return fmt.Sprintf("%x", h)
}

func rgbColor(c string) string {
	if c == "" {
		return ""
	}
	rgba := parseHexColor(c, color.RGBA{255, 255, 255, 255})
	return fmt.Sprintf("rgb(%d, %d, %d)", rgba.R, rgba.G, rgba.B)
}

func parseHexColor(s string, fallback color.RGBA) color.RGBA {
	if len(s) == 0 || s[0] != '#' {
		return fallback
	}
	s = s[1:]
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return fallback
	}
	var c color.RGBA
	c.A = 255
	if _, err := fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B); err != nil {
		return fallback
	}
	return c
}

func fetchImage(url string) ([]byte, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch image: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
