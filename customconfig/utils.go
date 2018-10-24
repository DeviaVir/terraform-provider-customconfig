package customconfig

import (
  "crypto/sha256"
  "encoding/hex"
)


func hash(s string) string {
  sha := sha256.Sum256([]byte(s))
  return hex.EncodeToString(sha[:])
}
