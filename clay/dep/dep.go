package dep

type DepTrackr struct {
	Dirpath string
}

func NewDepTrackr(dirpath string) *DepTrackr {
	return &DepTrackr{
		Dirpath: dirpath,
	}
}

func (d *DepTrackr) Load() error {

	return nil
}

func (d *DepTrackr) Save() error {

	return nil
}

type Digest struct {
	Hash [20]byte
}

type ItemToAdd struct {
	IdData     []byte
	IdDigest   Digest
	ItemData   []byte
	ItemDigest Digest
	Flags      uint32
}

type State int

const (
	StateUpToDate State = iota
	StateOutOfDate
)

func (d *DepTrackr) Add(item ItemToAdd, deps []ItemToAdd) (State, error) {

	return StateOutOfDate, nil
}

func (d *DepTrackr) Remove(item ItemToAdd) error {

	return nil
}

func (d *DepTrackr) Query(item ItemToAdd, deps []ItemToAdd) (State, error) {

	return StateOutOfDate, nil
}
