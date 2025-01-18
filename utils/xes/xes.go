package xes

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Event represents a parsed event from the XES format
type Event struct {
	CaseID     string                 `json:"case_id"`
	ActivityID string                 `json:"activity_id"`
	Timestamp  string                 `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
}

// XesWrapper represents the custom wrapper element
type XesWrapper struct {
	XMLName xml.Name `xml:"org.deckfour.xes.model.impl.XTraceImpl"`
	Log     XesLog   `xml:"log"`
}

// XesLog represents the root structure of the XES document inside the wrapper
type XesLog struct {
	XMLName xml.Name `xml:"log"`
	Trace   XesTrace `xml:"trace"`
}

// XesTrace represents a trace in the XES document
type XesTrace struct {
	Events  []XesEvent  `xml:"event"`
	Strings []XesString `xml:"string"` // Stores string fields for trace (like CaseID)
}

// TODO: missing flowattributes
// XesEvent represents an event within a trace
type XesEvent struct {
	StringAttributes  []XesString  `xml:"string"` // Stores string fields for the event (like ActivityID)
	IntegerAttributes []XesInt     `xml:"int"`
	BooleanAttributes []XesBoolean `xml:"boolean"`
	FloatAttributes   []XesFloat   `xml:"float"`
	Timestamp         []XesDate    `xml:"date"` // Stores the timestamp
}

// XesString represents a key-value pair for string fields
type XesString struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}
type XesInt struct {
	Key   string `xml:"key,attr"`
	Value int    `xml:"value,attr"`
}
type XesBoolean struct {
	Key   string `xml:"key,attr"`
	Value bool   `xml:"value,attr"`
}
type XesFloat struct {
	Key   string  `xml:"key,attr"`
	Value float64 `xml:"value,attr"`
}

// XesDate represents a key-value pair for date fields
type XesDate struct {
	Key   string    `xml:"key,attr"`
	Value time.Time `xml:"value,attr"`
}

// ParseXes parses an XES formatted string and returns an Event object
func ParseXes(xesString string) (*Event, error) {
	var wrapper XesWrapper
	err := xml.Unmarshal([]byte(xesString), &wrapper)
	if err != nil {
		return nil, fmt.Errorf("error parsing XES: %v", err)
	}

	// Extract CaseID from trace level string fields
	var caseID string
	for _, str := range wrapper.Log.Trace.Strings {
		if str.Key == "concept:name" {
			caseID = str.Value
			break
		}
	}
	if len(wrapper.Log.Trace.Events) == 0 {
		return nil, fmt.Errorf("no events found in trace")
	}

	firstEvent := wrapper.Log.Trace.Events[0]

	// Extract ActivityID from event level string fields
	var activityID string

	attr_map := make(map[string]interface{})
	for _, str := range firstEvent.StringAttributes {
		if str.Key == "concept:name" {
			activityID = str.Value
		} else {
			attr_map[str.Key] = str.Value
		}
	}

	for _, i := range firstEvent.IntegerAttributes {
		attr_map[i.Key] = i.Value
	}
	for _, b := range firstEvent.BooleanAttributes {
		attr_map[b.Key] = b.Value
	}
	for _, f := range firstEvent.FloatAttributes {
		attr_map[f.Key] = f.Value
	}
	var timestamp string
	for _, d := range firstEvent.Timestamp {
		if d.Key == "time:timestamp" {
			timestamp = d.Value.Format(time.RFC3339)
		} else {
			attr_map[d.Key] = d.Value.Format(time.RFC3339)
		}
	}
	// Create Event object
	event := &Event{
		CaseID:     caseID,
		ActivityID: activityID,
		Timestamp:  timestamp,
		Attributes: attr_map,
	}

	return event, nil
}

// Stringify generates a full XES XML string from an Event
func Stringify(event Event) (string, error) {
	// Convert the Event into an XesEvent structure
	xesEvent := XesEvent{}

	// Map the string attributes from the Event
	for key, value := range event.Attributes {
		switch v := value.(type) {
		case string:
			xesEvent.StringAttributes = append(xesEvent.StringAttributes, XesString{Key: key, Value: v})
		case int:
			xesEvent.IntegerAttributes = append(xesEvent.IntegerAttributes, XesInt{Key: key, Value: v})
		case bool:
			xesEvent.BooleanAttributes = append(xesEvent.BooleanAttributes, XesBoolean{Key: key, Value: v})
		case float64:
			xesEvent.FloatAttributes = append(xesEvent.FloatAttributes, XesFloat{Key: key, Value: v})
		case time.Time:
			xesEvent.Timestamp = append(xesEvent.Timestamp, XesDate{Key: key, Value: v})
		default:
			return "", fmt.Errorf("unsupported attribute type for key '%s': %T", key, v)
		}
	}

	// Add the CaseID and ActivityID as specific string attributes
	xesEvent.StringAttributes = append(xesEvent.StringAttributes,
		XesString{Key: "concept:name", Value: event.ActivityID})

	// Create a trace with the CaseID as a string field
	trace := XesTrace{
		Strings: []XesString{
			{Key: "concept:name", Value: event.CaseID},
		},
		Events: []XesEvent{xesEvent},
	}

	// Create the wrapper and log structure
	wrapper := XesWrapper{
		Log: XesLog{
			Trace: trace,
		},
	}

	// Serialize the wrapper to XML
	xmlData, err := xml.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error serializing XES: %v", err)
	}

	return string(xmlData), nil
}

func main() {
	// Example Event
	event := Event{
		CaseID:     "Case123",
		ActivityID: "ActivityStart",
		Timestamp:  time.Now().Format(time.RFC3339),
		Attributes: map[string]interface{}{
			"custom:string":  "exampleValue",
			"custom:int":     42,
			"custom:bool":    true,
			"custom:float":   3.14,
			"time:timestamp": time.Now(),
		},
	}

	// Generate the XES XML string
	xesXML, err := Stringify(event)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the result
	fmt.Println(xesXML)
}

/*
<org.deckfour.xes.model.impl.XTraceImpl>
  <log
        openxes.version="1.0RC7"
        xes.features="nested-attributes"
        xes.version="1.0"
        xmlns="http://www.xes-standard.org/">
    <trace>
      <string key="concept:name" value="case_3"/>
      <event>
        <string key="concept:name" value="Activity G"/>
        <date key="time:timestamp" value="2024-10-03T16:27:33.682+02:00"/>
        <string key="organization" value="organization_A"/>
      </event>
    </trace>
  </log>
</org.deckfour.xes.model.impl.XTraceImpl>
*/
