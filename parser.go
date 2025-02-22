package qp

import "encoding/json"

func Unmarshal(data []byte) (where *QueryParams, err error) {
	where = &QueryParams{}
	err = where.Unmarshal(data)
	return
}

func (qp *QueryParam) Marshal() (data []byte, err error) {
	return json.Marshal(qp)
}

func (qp *QueryParam) Unmarshal(data []byte) (err error) {
	if data != nil {
		if len(data) != 0 {
			return json.Unmarshal(data, &qp)
		}
	}
	return
}

func (qp *QueryParams) Marshal() (data []byte, err error) {
	return json.Marshal(qp)
}

func (qp *QueryParams) Unmarshal(data []byte) (err error) {
	if data != nil {
		if len(data) != 0 {
			return json.Unmarshal(data, &qp)
		}
	}
	return
}
