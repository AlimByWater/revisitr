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
	"golang.org/x/crypto/pkcs12"

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

	privateKey, certificate, err := pkcs12.Decode(certData, password)
	if err != nil {
		return nil, fmt.Errorf("pkcs12 decode: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("certificate private key is not RSA")
	}

	// Extract additional CA certificates from the PKCS12 bundle
	var caCerts []*x509.Certificate
	if pemBlocks, err := pkcs12.ToPEM(certData, password); err == nil {
		for _, block := range pemBlocks {
			if block.Type == "CERTIFICATE" {
				if c, err := x509.ParseCertificate(block.Bytes); err == nil {
					if !c.Equal(certificate) {
						caCerts = append(caCerts, c)
					}
				}
			}
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

	// Add images
	var iconData []byte
	if cfg.Design.LogoURL != "" {
		logoData, err := fetchImage(cfg.Design.LogoURL)
		if err == nil {
			manifest["logo.png"] = sha1Hex(logoData)
			_ = logoData // stored below
		}
	}

	// Generate icon
	iconData, err = g.generateIcon(cfg.Design.BackgroundColor)
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
	return g.zipPass(passJSON, iconData, qrData, signature, cfg.Design.LogoURL)
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

	// DER-encode authenticated attributes (for signing)
	aaDER, err := asn1.Marshal(struct {
		Attrs []pkcs7Attribute `asn1:"set,implicit,tag:0"`
	}{Attrs: attrs})
	if err != nil {
		return nil, fmt.Errorf("marshal auth attrs: %w", err)
	}

	// RSA-SHA1 sign the authenticated attributes DER
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, aaDER)
	if err != nil {
		return nil, fmt.Errorf("rsa sign: %w", err)
	}

	// Build certificates SET
	var certDERs []asn1.RawValue
	for _, c := range certs {
		certDERs = append(certDERs, asn1.RawValue{FullBytes: c.Raw})
	}

	certSet, err := asn1.Marshal(struct {
		Certs []asn1.RawValue `asn1:"set"`
	}{Certs: certDERs})
	if err != nil {
		return nil, fmt.Errorf("marshal cert set: %w", err)
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
		Certificates: asn1.RawValue{FullBytes: certSet, Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true},
		SignerInfos:  []pkcs7SignerInfo{signerInfo},
	}

	sdDER, err := asn1.Marshal(signedData)
	if err != nil {
		return nil, fmt.Errorf("marshal signedData: %w", err)
	}

	// Build top-level ContentInfo
	ci := pkcs7ContentInfo{
		ContentType: oidSignedData,
		Content:     asn1.RawValue{FullBytes: sdDER, Class: asn1.ClassContextSpecific, Tag: 0, IsCompound: true},
	}

	return asn1.Marshal(ci)
}

func (g *PassGenerator) zipPass(passJSON, iconData, qrData, signature []byte, logoURL string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// pass.json
	f, err := w.Create("pass.json")
	if err != nil {
		return nil, fmt.Errorf("zip pass.json: %w", err)
	}
	if _, err := f.Write(passJSON); err != nil {
		return nil, fmt.Errorf("write pass.json: %w", err)
	}

	// icon.png
	f, err = w.Create("icon.png")
	if err != nil {
		return nil, fmt.Errorf("zip icon.png: %w", err)
	}
	if _, err := f.Write(iconData); err != nil {
		return nil, fmt.Errorf("write icon.png: %w", err)
	}

	// logo.png (if available)
	if logoURL != "" {
		logoData, err := fetchImage(logoURL)
		if err == nil {
			f, err = w.Create("logo.png")
			if err != nil {
				return nil, fmt.Errorf("zip logo.png: %w", err)
			}
			if _, err := f.Write(logoData); err != nil {
				return nil, fmt.Errorf("write logo.png: %w", err)
			}
		}
	}

	// signature
	f, err = w.Create("signature")
	if err != nil {
		return nil, fmt.Errorf("zip signature: %w", err)
	}
	if _, err := f.Write(signature); err != nil {
		return nil, fmt.Errorf("write signature: %w", err)
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
