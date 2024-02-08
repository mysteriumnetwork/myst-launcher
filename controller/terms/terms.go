package terms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mysteriumnetwork/myst-launcher/model"
	"golang.org/x/net/context/ctxhttp"
)

func CheckTermsAgreement(model *model.UIModel) {

	fetch := func(res *[]byte) error {

		url := "https://raw.githubusercontent.com/mysteriumnetwork/node/master/TERMS_EXIT_NODE.md"
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		// pin to API version 3 to avoid breaking our structs
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request faild with %v (%v)", resp.StatusCode, resp.Status)
		}

		*res, err = io.ReadAll(resp.Body)
		return err
	}

	if model.Config.AgreementConsentDate.IsZero() {
		// download agreement
		doc := make([]byte, 0)
		fetch(&doc)
		res := strings.NewReplacer("\n", "\r\n").Replace(string(doc))
		model.UIBus.Publish("show-agreement", string(res))

		model.UIBus.SubscribeOnce("accept-agreement", func() {
			model.Config.AgreementConsentDate = time.Now()
			model.Config.Save()
		})
	}
}
