package gormimpl

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

func MustCheckErrors(db *gorm.DB) {
	err := CheckErrors(db)
	if err != nil {
		panic(err)
	}
}

func CheckErrors(db *gorm.DB) error {
	errs := db.GetErrors()
	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return fmt.Errorf("gorm error: %v", errs[0])
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "gorm error: (total %d)\n", len(errs))
	for i, err := range errs {
		fmt.Fprintf(&buf, "error %d: %v\n", i+1, err)
	}

	return errors.New(buf.String())
}
