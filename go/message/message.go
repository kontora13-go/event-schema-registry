package message

import "time"

// EventMessage - основная структура event-сообщения
// schemagen:
//
//	#is_event_message
type EventMessage struct {
	Meta *EventMeta `json:"meta" schema:"meta,required"`
	Data any        `json:"data" schema:"data,required,event_data"`
}

// EventMeta - метеданные event-сообщения
type EventMeta struct {
	TraceId      string    `json:"trace_id" schema:"trace_id,required"`
	EventId      string    `json:"event_id" schema:"event_id,required"`
	EventName    string    `json:"event_name" schema:"event_name,required"`
	EventVersion string    `json:"event_version" schema:"event_version,required"`
	EventTime    time.Time `json:"event_time" schema:"event_time,required"`
	Producer     string    `json:"producer" schema:"producer,required"`
}
