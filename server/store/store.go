package store

type Store struct {
	messages chan<- message
}

type message interface {
	Filename() string
}

type putMessage struct {
	filename string
	Contents []byte
}

type getMessage struct {
	filename string
	reply    chan<- getReply
}

type getReply struct {
	contents []byte
	prs      bool
}

func (msg getMessage) Filename() string {
	return msg.filename
}

func (msg putMessage) Filename() string {
	return msg.filename
}

func New() Store {
	messages := make(chan message)
	go func() {
		files := make(map[string][]byte)
		for {
			msg := <-messages
			switch msg.(type) {
			case getMessage:
				get := msg.(getMessage)
				f, prs := files[msg.Filename()]
				get.reply <- getReply{contents: f, prs: prs}
			case putMessage:
				put := msg.(putMessage)
				files[msg.Filename()] = put.Contents
			}
		}
	}()
	return Store{messages: messages}
}

func (store *Store) Put(filename string, contents []byte) {
	msg := putMessage{filename: filename, Contents: contents}
	store.messages <- msg
}

func (store *Store) Get(filename string) ([]byte, bool) {
	reply := make(chan getReply)
	msg := getMessage{filename: filename, reply: reply}
	store.messages <- msg
	r := <-reply
	return r.contents, r.prs
}
