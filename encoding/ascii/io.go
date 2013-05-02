package ascii

import (
	"fmt"
	"io"

	"github.com/go-hep/hepmc"
)

const (
	genevent_start      = "HepMC::IO_GenEvent-START_EVENT_LISTING"
	ascii_start         = "HepMC::IO_Ascii-START_EVENT_LISTING"
	extendedascii_start = "HepMC::IO_ExtendedAscii-START_EVENT_LISTING"

	genevent_end      = "HepMC::IO_GenEvent-END_EVENT_LISTING"
	ascii_end         = "HepMC::IO_Ascii-END_EVENT_LISTING"
	extendedascii_end = "HepMC::IO_ExtendedAscii-END_EVENT_LISTING"

	pdt_start               = "HepMC::IO_Ascii-START_PARTICLE_DATA"
	extendedascii_pdt_start = "HepMC::IO_ExtendedAscii-START_PARTICLE_DATA"
	pdt_end                 = "HepMC::IO_Ascii-END_PARTICLE_DATA"
	extendedascii_pdt_end   = "HepMC::IO_ExtendedAscii-END_PARTICLE_DATA"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (enc *Encoder) Encode(evt *hepmc.Event) error {
	var err error

	_, err = fmt.Fprintf(enc.w, "%s\n", genevent_start)
	return err
}

// EOF
