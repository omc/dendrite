package dendrite

import (
	"io"
)

type Destinations []*Destination

type Destination struct {
	Encoder Encoder
	RW      io.ReadWriter
}

func (dests *Destinations) Consume(ch chan Record) {
  for {
    rec := <-ch
    
    if rec == nil{
      break
    } else {
      for _, dest := range *dests {
        dest.Encoder.Encode(rec, dest.RW)
      }
    }
  } 
}

func (dests *Destinations) Reader() io.Reader {
  var readers = make([]io.Reader, 0)
  for _, dest := range *dests {
    readers = append(readers, dest.RW)
  }
  return NewAnyReader(readers)
}

func NewDestinations() Destinations {
	return make([]*Destination, 0)
}

func NewDestination(config DestinationConfig) (*Destination, error) {
  var err error = nil
  dest := new(Destination)
  
  dest.RW, err = NewReadWriter(config.Url)
  if(err != nil) {
    return nil, err
  }
  
  dest.Encoder, err = NewEncoder(config.Url)
  if(err != nil) {
    return nil, err
  }
  
  return dest, nil
}