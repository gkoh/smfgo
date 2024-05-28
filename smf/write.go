package smf

import (
	"encoding/xml"
	"fmt"
)

// GenerateXML marshals the service bundle into an XML multiline string.
func (b *ServiceBundle) GenerateXML() (string, error) {
	output, err := xml.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s%s", Header, DocType, string(output)), nil
}
