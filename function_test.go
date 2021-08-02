package tracker

import (
	"reflect"
	"testing"
)

func TestCreateMessage(t *testing.T) {
	want := `
Name: Shree Janardhan Swami Hospital
Area:  Kopargaon
Pincode:  0
Type:  
Fee:  
Date:  
Age Limit:  0
Vaccine:  **
        
Dose1:  *0*`
	data := DistSessions{
		Name:    "Shree Janardhan Swami Hospital",
		Address: "Kopargaon",
	}
	got := CreateMessage(data)
	if !reflect.DeepEqual(want, got) {
		t.Error(got)
	}
}
