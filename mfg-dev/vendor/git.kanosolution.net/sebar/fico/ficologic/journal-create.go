package ficologic

import "git.kanosolution.net/kano/dbflex/orm"

type JournalCreate interface {
	Create() (orm.DataModel, error)
}
