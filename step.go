package sipengine

type SIPStep func(message *Message) error