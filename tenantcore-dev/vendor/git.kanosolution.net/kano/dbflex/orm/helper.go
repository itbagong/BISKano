package orm

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"github.com/ariefdarmawan/reflector"
	"github.com/sebarcode/codekit"
)

func ensureFK(hub dbflex.IConnection, dm DataModel, fieldID, sourceTableName, sourceField string,
	create bool, seqNumID string, refMap codekit.M, defaultFields ...string) error {
	rf := reflector.From(dm)
	keyValue, e := rf.Get(fieldID)
	if e != nil {
		return errors.New("fkErr: " + sourceTableName + ", " + e.Error())
	}

	if keyValue != "" {
		flush := false
		mRef := codekit.M{}
		cmdGetRef := dbflex.From(sourceTableName).Where(dbflex.Eq(sourceField, keyValue)).Select().Take(1)
		if e = hub.Cursor(cmdGetRef, nil).Fetch(&mRef).Close(); e != nil {
			if create {
				cmdSave := dbflex.From(sourceTableName).Save()
				if seqNumID == "" {
					mRef.Set(sourceField, keyValue)
				} else {
					numseq := codekit.M{}
					cmdGet := dbflex.From("KNSequences").Where(dbflex.Eq("_id", seqNumID)).Select()
					if e := hub.Cursor(cmdGet, nil).Fetch(&numseq).Close(); e != nil {
						return e
					}
					lastNo := numseq.GetInt("LastNo") + 1
					format := numseq.GetString("Format")
					if format == "" {
						format = "%08d"
					}
					fmtLastNo := fmt.Sprintf(format, lastNo)
					mRef.Set("_id", fmtLastNo)
					mRef.Set("Enable", true)
					numseq.Set("LastNo", lastNo)
					cmdUpdate := dbflex.From("KNSequences").Where(dbflex.Eq("_id", seqNumID)).Update("LastNo")
					if _, e := hub.Execute(cmdUpdate, codekit.M{}.Set("data", numseq)); e != nil {
						return e
					}
					rf.Set(fieldID, fmtLastNo)
					flush = true
				}
				for _, def := range defaultFields {
					mRef.Set(def, keyValue)
				}
				if _, eSave := hub.Execute(cmdSave, codekit.M{}.Set("data", mRef)); eSave != nil {
					return errors.New("createFK: " + sourceTableName + ", " + eSave.Error())
				}
			} else {
				return errors.New("missingFK: " + sourceTableName)
			}
		}

		if refMap != nil {
			for field, source := range refMap {
				rf.Set(field, mRef.GetString(source.(string)))
			}
			flush = true
		}

		if flush {
			rf.Flush()
		}
	}

	return nil
}

func ensureEmptyFK(hub dbflex.IConnection, dm DataModel, fieldID, refTableName, refField string, autoDel bool) error {
	sourceM, e := codekit.ToM(dm)
	if e != nil {
		return errors.New("fkErr: " + refTableName)
	}
	keyValue := sourceM.GetString(fieldID)

	if keyValue != "" {
		cmd := dbflex.From(refTableName).Where(dbflex.Eq(refField, keyValue)).Select().Take(1)
		refM := codekit.M{}
		if e = hub.Cursor(cmd, nil).Fetch(&refM).Close(); e == nil {
			if !autoDel {
				return errors.New("fkIsNotEmpty: " + refTableName)
			}

			cmdDel := dbflex.From(refTableName).Where(dbflex.Eq(refField, keyValue)).Delete()
			if _, e := hub.Execute(cmdDel, nil); e != nil {
				return errors.New("fkAutoDeleteErr: " + refTableName + ", " + e.Error())
			}
		}
	}
	return nil
}
