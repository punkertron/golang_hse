//go:build !solution

package retryupdate

import (
	"errors"

	"github.com/gofrs/uuid"
	"gitlab.com/slon/shad-go/03-Interfaces/retryupdate/kvapi"
)

func UpdateValue(
	c kvapi.Client,
	key string,
	updateFn func(oldValue *string) (newValue string, err error),
) error {
GetLoop:
	for {
		resp, err := c.Get(&kvapi.GetRequest{Key: key})

		var authError *kvapi.AuthError
		if errors.As(err, &authError) {
			return err
		}

		if err != nil && !errors.Is(err, kvapi.ErrKeyNotFound) {
			continue GetLoop
		}

	UpdateValLoop:
		for {
			var newValue string

			var oldVersion uuid.UUID

			if resp == nil {
				newValue, err = updateFn(nil)
			} else {
				newValue, err = updateFn(&resp.Value)
				oldVersion = resp.Version
			}

			if err != nil {
				return err
			}

			setRequest := &kvapi.SetRequest{
				Key:        key,
				Value:      newValue,
				OldVersion: oldVersion,
				NewVersion: uuid.Must(uuid.NewV4()),
			}

			for {
				_, err = c.Set(setRequest)
				if err == nil {
					return nil
				}

				if errors.As(err, &authError) {
					return err
				}

				if errors.Is(err, kvapi.ErrKeyNotFound) {
					resp = nil
					continue UpdateValLoop
				}

				var conflictError *kvapi.ConflictError
				if errors.As(err, &conflictError) {
					if conflictError.ExpectedVersion == setRequest.NewVersion {
						return nil
					}

					continue GetLoop
				}
			}
		}
	}
}
