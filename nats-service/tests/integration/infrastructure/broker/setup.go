package broker

import "testing"

// SetupTestContainer initializes the TestContainer.
func SetupTestContainer(t *testing.T) *TestContainer {
	container := NewTestContainer()
	client := container.NatsClient.Get()

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	})

	return container
}
