package shelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

func PostJournal(journal ficologic.JournalPosting, userid string, op string, text string) (*tenantcoremodel.PreviewReport, error) {
	var (
		res *tenantcoremodel.PreviewReport
	)

	err := journal.ExtractHeader()
	if err != nil {
		return res, err
	}

	pp := journal.PostingProfile()
	if pp == nil {
		return res, errors.New("missing: posting profile")
	}

	err = journal.ExtractLines()
	if err != nil {
		return res, err
	}

	if err = journal.Validate(); err != nil {
		return res, fmt.Errorf("validate: %s", err.Error())
	}

	err = journal.Calculate()
	if err != nil {
		return res, fmt.Errorf("calculate: %s", err.Error())
	}

	preview := journal.Preview()

	switch op {
	case string(shemodel.PostOpPreview):
		return preview, nil

	case string(shemodel.PostOpSubmit):
		_, err = journal.Submit()
		if err == nil && journal.Status() == "READY" && pp.DirectPosting {
			if err = journal.Post(); err != nil {
				return preview, fmt.Errorf("posting fail: %s", err.Error())
			}
		}
		return preview, err

	case string(shemodel.PostOpApprove), string(shemodel.PostOpReject):
		_, err = journal.Approve(op, text)
		if err == nil && journal.Status() == "READY" && pp.DirectPosting {
			if err = journal.Post(); err != nil {
				return preview, fmt.Errorf("posting fail: %s", err.Error())
			}
		}
		return preview, err

	case string(shemodel.PostOpPost):
		err = journal.Post()
		return preview, err

	default:
		return res, fmt.Errorf("op %s is not known", op)
	}
}
