// SPDX-FileCopyrightText: Copyright 2021 The go-pexec Authors
// SPDX-License-Identifier: BSD-3-Clause

package pexec

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEventType(t *testing.T) {
	for _, eventType := range allEventTypes {
		t.Run(eventType.String(), func(t *testing.T) {
			data, err := eventType.MarshalJSON()
			if err != nil {
				t.Fatalf("could not json marshal: %v", err)
			}

			var unmarshalledEventType EventType
			if err := (&unmarshalledEventType).UnmarshalJSON(data); err != nil {
				t.Fatalf("could not json marshal: %v", err)
			}

			// require.NoError(t, (&unmarshalledEventType).UnmarshalJSON(data))
			if diff := cmp.Diff(eventType, unmarshalledEventType); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
			// assert.Equal(t, eventType, unmarshalledEventType)

			data, err = eventType.MarshalText()
			if err != nil {
				t.Fatalf("could not json marshal: %v", err)
			}
			// require.NoError(t, err)

			if err := (&unmarshalledEventType).UnmarshalJSON(data); err != nil {
				t.Fatalf("could not json marshal: %v", err)
			}
			// require.NoError(t, (&unmarshalledEventType).UnmarshalText(data))

			if diff := cmp.Diff(eventType, unmarshalledEventType); diff != "" {
				t.Fatalf("(-want +got):\n%s", diff)
			}
			// assert.Equal(t, eventType, unmarshalledEventType)
		})
	}
}
