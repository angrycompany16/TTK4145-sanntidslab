package distribute

var (
	ThisBackup Backup
)

type Backup struct {
	records []Record
}

// I'm disintegrating

// A struct that contains a backup of a single request in an elevator with a
// specified id
type Record struct {
	Request ElevatorRequest `json:"Request"`
	Id      string          `json:"Id"`
}

func InitBackup() {
	ThisBackup.records = make([]Record, 0)
}

// TODO: Just take a Record as parameter stupid
// or...?
func (b *Backup) AddRecord(request ElevatorRequest, id string) {
	b.records = append(b.records, Record{Request: request, Id: id})
}

// TODO: Error handling if invalid id?
func (b *Backup) GetRequests(id string) []ElevatorRequest {
	res := make([]ElevatorRequest, 0)
	for _, record := range b.records {
		if record.Id == id {
			res = append(res, record.Request)
		}
	}
	return res
}
