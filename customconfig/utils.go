package customconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func retry(retryFunc func() error, seconds int) error {
	return retryTime(retryFunc, seconds)
}

func retryTime(retryFunc func() error, seconds int) error {
	wait := 1
	return resource.Retry(time.Duration(seconds)*time.Second, func() *resource.RetryError {
		err := retryFunc()
		if err == nil {
			return nil
		}

		rand.Seed(time.Now().UnixNano())
		randomNumberMiliseconds := rand.Intn(1001)

		// Deal with a broken vault
		if strings.Contains(fmt.Sprintf("%s", err), "internal error") {
			log.Printf("[DEBUG] Retrying internal error response from API")
			time.Sleep(time.Duration(wait)*time.Second + time.Duration(randomNumberMiliseconds))
			wait = wait * 2
			return resource.RetryableError(err)
		}
		if strings.Contains(fmt.Sprintf("%s", err), "error talking to Vault") {
			log.Printf("[DEBUG] Retrying service unavailable from API")
			time.Sleep(time.Duration(wait)*time.Second + time.Duration(randomNumberMiliseconds))
			wait = wait * 2
			return resource.RetryableError(err)
		}
		if strings.Contains(fmt.Sprintf("%s", err), "error reading from Vault") {
			log.Printf("[DEBUG] Retrying due to eventual consistency")
			time.Sleep(time.Duration(wait)*time.Second + time.Duration(randomNumberMiliseconds))
			wait = wait * 2
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
}
