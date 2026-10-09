package qp

import "testing"

func TestT(t *testing.T) {
	qp, err := QueryFormat("[user_id, fio][~*]", "test||[login, username][>]=temp")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(qp)
}
