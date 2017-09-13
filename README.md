# datastore-serializer

A serializer that works in conjonction with [Google Cloud Datastore](https://cloud.google.com/datastore/docs/concepts/overview) to serialize pointer relationship. This is useful if you go the denormalization way.

## Installation

```
go get github.com/marcgiovannoni/datastore-serializer
```

## Usage

This serializer has been written so working with [Google JSONAPI](https://github.com/google/jsonapi) and [Google Cloud Datastore](https://cloud.google.com/datastore/docs/concepts/overview) is easier.

The use of jsonapi is not mandatory, it will work for simple relation serialization.

### Requirements

`ID` attribute has to be an encoded datastore.Key

```golang

type Post struct {
    ID       string     `datastore:"-" jsonapi:"primary,post"`
    Text     string     `datastore:"text,noindex" jsonapi:"attr,text"`
    Comments []*Comment `datastore:"-" jsonapi:"comments,text" serializer:"relation,comments"`
}

type Comment struct {
    ID   string `datastore:"-" jsonapi:"primary,comment"`
    Text string `datastore:"text,noindex" jsonapi:"attr,text"`
}

func (me *Post) Load(ps []datastore.Property) error {
    return serializer.LoadEntity(me, ps)
}

func (me *Post) Save() ([]datastore.Property, error) {
    ps, err := serializer.SaveEntity(me)
    if err != nil {
        return nil, err
    }
    return ps, nil
}

func (me *Comment) Load(ps []datastore.Property) error {
    return serializer.LoadEntity(me, ps)
}

func (me *Comment) Save() ([]datastore.Property, error) {
    ps, err := serializer.SaveEntity(me)
    if err != nil {
        return nil, err
    }
    return ps, nil
}
```

This will be serialized in the datastore as:

```
text comments.id comments.text
```
