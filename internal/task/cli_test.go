package task

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/YuriLeonel/task-manager/internal/storage"
	"github.com/YuriLeonel/task-manager/pkg/models"
)

// CLIMockStorage implements storage.Storage for CLI testing
type CLIMockStorage struct {
	tasks map[string]*models.Task
}

func NewCLIMockStorage() *CLIMockStorage {
	return &CLIMockStorage{
		tasks: make(map[string]*models.Task),
	}
}

func (m *CLIMockStorage) SaveTask(ctx context.Context, task *models.Task) error {
	m.tasks[task.ID] = task
	return nil
}

func (m *CLIMockStorage) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if task, ok := m.tasks[id]; ok {
		return task, nil
	}
	return nil, nil
}

func (m *CLIMockStorage) GetAllTasks(ctx context.Context) ([]*models.Task, error) {
	tasks := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *CLIMockStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	if _, ok := m.tasks[task.ID]; ok {
		m.tasks[task.ID] = task
		return nil
	}
	return nil
}

func (m *CLIMockStorage) DeleteTask(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}

func (m *CLIMockStorage) CreateBackup(ctx context.Context) error {
	return nil
}

func (m *CLIMockStorage) RestoreBackup(ctx context.Context, backupID string) error {
	return nil
}

func (m *CLIMockStorage) ListBackups(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *CLIMockStorage) DeleteBackup(ctx context.Context, backupID string) error {
	return nil
}

// captureOutput captures stdout and stderr during test execution
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	f()

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return <-outC
}

// mockStdin simulates user input during tests
type mockStdin struct {
	*bytes.Buffer
}

func newMockStdin(input string) *os.File {
	r, w, _ := os.Pipe()
	go func() {
		w.Write([]byte(input))
		w.Close()
	}()
	return r
}

// mockInput simulates user input
type mockInput struct {
	input string
}

func (m *mockInput) Read(p []byte) (n int, err error) {
	if m.input == "" {
		return 0, io.EOF
	}
	n = copy(p, m.input)
	m.input = m.input[n:]
	return n, nil
}

type mockReader struct {
	inputs []string
	index  int
}

func (m *mockReader) ReadString(delim byte) (string, error) {
	if m.index >= len(m.inputs) {
		return "", nil
	}
	input := m.inputs[m.index]
	m.index++
	return input + "\n", nil
}

func TestCLI(t *testing.T) {
	mockStorage, err := storage.NewMemoryStorage(100, "test_backups", 5)
	if err != nil {
		t.Fatalf("Failed to create mock storage: %v", err)
	}
	service := NewService(mockStorage)
	var output bytes.Buffer
	mockInputs := []string{
		"1",          // Add task
		"Test task",  // Task description
		"2",          // List tasks
		"3",          // Mark as completed
		"1",          // Task number
		"4",          // Set priority
		"1",          // Task number
		"2",          // Medium priority
		"5",          // Set due date
		"1",          // Task number
		"2023-12-31", // Due date
		"7",          // Add tag
		"1",          // Task number
		"test-tag",   // Tag name
		"8",          // Remove tag
		"1",          // Task number
		"1",          // Tag number
		"9",          // Set progress
		"1",          // Task number
		"50",         // Progress value
		"11",         // Exit
	}

	input := strings.Join(mockInputs, "\n") + "\n"
	cli := &CLI{
		service: service,
		reader:  bufio.NewReader(strings.NewReader(input)),
		writer:  &output,
	}

	if err := cli.Run(); err != nil {
		t.Fatalf("CLI run failed: %v", err)
	}

	outputStr := output.String()
	expectedOutputs := []string{
		"Task added successfully!",
		"Test task",
		"Task marked as completed!",
		"Task priority updated!",
		"Task due date updated!",
		"Tag added successfully!",
		"Tag removed successfully!",
		"Progress set to 50%",
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}
}

func TestCLIInvalidInput(t *testing.T) {
	mockStorage, err := storage.NewMemoryStorage(100, "test_backups", 5)
	if err != nil {
		t.Fatalf("Failed to create mock storage: %v", err)
	}
	service := NewService(mockStorage)
	var output bytes.Buffer
	mockInputs := []string{
		"1",            // Add task
		"Test task",    // Task description
		"3",            // Mark as completed
		"999",          // Invalid task number
		"4",            // Set priority
		"1",            // Task number
		"999",          // Invalid priority
		"5",            // Set due date
		"1",            // Task number
		"invalid-date", // Invalid date
		"9",            // Set progress
		"1",            // Task number
		"150",          // Invalid progress value
		"11",           // Exit
	}

	input := strings.Join(mockInputs, "\n") + "\n"
	cli := &CLI{
		service: service,
		reader:  bufio.NewReader(strings.NewReader(input)),
		writer:  &output,
	}

	if err := cli.Run(); err != nil {
		t.Fatalf("CLI run failed: %v", err)
	}

	outputStr := output.String()
	expectedErrors := []string{
		"invalid task number",
		"invalid priority level",
		"invalid date format",
		"progress must be between 0 and 100",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected output to contain error %q, but it didn't", expected)
		}
	}
}
