import useSWR, { useSWRConfig, type SWRConfiguration } from 'swr'
import useSWRMutation from 'swr/mutation'

/**
 * Wrapper around useSWR that returns `isError` (boolean) for backward compatibility
 * with existing page components that used TanStack Query's `isError`.
 */
export function useApiQuery<T>(
  key: string | null,
  fetcher: () => Promise<T>,
  options?: SWRConfiguration<T>,
) {
  const result = useSWR<T>(key, fetcher, options)
  return {
    ...result,
    isError: !!result.error,
  }
}

/**
 * Wrapper around useSWRMutation that provides a backward-compatible interface
 * matching TanStack Query's useMutation return shape:
 *   - mutate(arg) / mutateAsync(arg) -> trigger(arg)
 *   - isPending -> isMutating
 *   - isError -> !!error
 *   - isSuccess -> data !== undefined && !error
 *
 * Cache invalidation is handled via `invalidateKeys` — all matching keys
 * are revalidated on success.
 */
export function useApiMutation<TData, TArg>(
  key: string,
  fn: (arg: TArg) => Promise<TData>,
  invalidateKeys?: string[],
) {
  const { mutate: globalMutate } = useSWRConfig()

  const result = useSWRMutation<TData, Error, string, TArg>(
    key,
    (_key: string, { arg }: { arg: TArg }) => fn(arg),
    {
      onSuccess: () => {
        invalidateKeys?.forEach((k) => {
          globalMutate(
            (cacheKey: string) =>
              typeof cacheKey === 'string' && cacheKey.startsWith(k),
            undefined,
            { revalidate: true },
          )
        })
      },
    },
  )

  return {
    ...result,
    mutate: result.trigger,
    mutateAsync: result.trigger,
    isPending: result.isMutating,
    isError: !!result.error,
    isSuccess: result.data !== undefined && !result.error,
  }
}
