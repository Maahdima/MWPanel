export function generateKeyPair() {
  const randomKey = () => Math.random().toString(36).substring(2, 34)
  return {
    privateKey: `priv_${randomKey()}`,
    publicKey: `pub_${randomKey()}`,
  }
}
