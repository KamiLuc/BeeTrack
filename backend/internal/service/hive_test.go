package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type mockApiaryMembershipReader struct {
	apiary *model.Apiary
	role   string
}

func (m *mockApiaryMembershipReader) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	if m.apiary == nil || m.apiary.ID != apiaryID {
		return nil, "", gorm.ErrRecordNotFound
	}
	return m.apiary, m.role, nil
}

type mockMultiApiaryReader struct {
	apiaries map[int64]*model.Apiary
	roles    map[int64]string
}

func (m *mockMultiApiaryReader) GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error) {
	a, ok := m.apiaries[apiaryID]
	if !ok {
		return nil, "", gorm.ErrRecordNotFound
	}
	return a, m.roles[apiaryID], nil
}

type mockHiveRepo struct {
	occupied      bool
	duplicateName bool
	created       *model.Hive
	hive          *model.Hive
	hives         []*model.Hive
	hivesByID     map[int64][]*model.Hive
	updated       *model.Hive
	deletedID     int64
	relocated     *model.Hive

	existsByNameApiaryID  int64
	existsByNameName      string
	existsByNameExcludeID int64
	existsByNameCalled    bool
}

func (m *mockHiveRepo) Create(ctx context.Context, h *model.Hive) error {
	h.ID = 1
	m.created = h
	return nil
}

func (m *mockHiveRepo) GetByIDAndApiaryID(ctx context.Context, hiveID, apiaryID int64) (*model.Hive, error) {
	if m.hive == nil || m.hive.ID != hiveID || m.hive.ApiaryID != apiaryID {
		return nil, gorm.ErrRecordNotFound
	}
	return m.hive, nil
}

func (m *mockHiveRepo) IsPositionOccupied(ctx context.Context, apiaryID int64, row, col int) (bool, error) {
	return m.occupied, nil
}

func (m *mockHiveRepo) ExistsByName(ctx context.Context, apiaryID int64, name string, excludeHiveID int64) (bool, error) {
	m.existsByNameCalled = true
	m.existsByNameApiaryID = apiaryID
	m.existsByNameName = name
	m.existsByNameExcludeID = excludeHiveID
	return m.duplicateName, nil
}

func (m *mockHiveRepo) ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error) {
	if m.hivesByID != nil {
		if hives, ok := m.hivesByID[apiaryID]; ok {
			return hives, nil
		}
		return []*model.Hive{}, nil
	}
	return m.hives, nil
}

func (m *mockHiveRepo) Move(ctx context.Context, hiveID int64, row, col int) error {
	if m.hive != nil && m.hive.ID == hiveID {
		m.hive.GridRow = row
		m.hive.GridCol = col
	}
	return nil
}

func (m *mockHiveRepo) Relocate(ctx context.Context, hiveID, newApiaryID int64, row, col int) error {
	if m.hive != nil && m.hive.ID == hiveID {
		m.hive.ApiaryID = newApiaryID
		m.hive.GridRow = row
		m.hive.GridCol = col
		m.relocated = m.hive
	}
	return nil
}

func (m *mockHiveRepo) Update(ctx context.Context, h *model.Hive) error {
	m.updated = h
	return nil
}

func (m *mockHiveRepo) Delete(ctx context.Context, hiveID int64) error {
	m.deletedID = hiveID
	return nil
}

func (m *mockHiveRepo) CreateDisease(ctx context.Context, d *model.HiveDisease) error {
	d.ID = 1
	return nil
}

func (m *mockHiveRepo) DeleteDisease(ctx context.Context, diseaseID, hiveID int64) error {
	return nil
}

func (m *mockHiveRepo) GetDiseaseByID(ctx context.Context, diseaseID, hiveID int64) (*model.HiveDisease, error) {
	return &model.HiveDisease{ID: diseaseID, HiveID: hiveID, Disease: "nosema"}, nil
}

func (m *mockHiveRepo) ListDiseasesByHiveID(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error) {
	return []*model.HiveDisease{}, nil
}

func (m *mockHiveRepo) ListDiseasesByHiveIDs(ctx context.Context, ids []int64) ([]*model.HiveDisease, error) {
	return []*model.HiveDisease{}, nil
}

func newTestHiveService() (*HiveService, *mockApiaryMembershipReader, *mockHiveRepo) {
	apiaryMock := &mockApiaryMembershipReader{}
	hiveMock := &mockHiveRepo{}
	svc := NewHiveService(apiaryMock, hiveMock)
	return svc, apiaryMock, hiveMock
}

func TestListHives_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hives = []*model.Hive{
		{ID: 1, Name: "Hive A", GridRow: 0, GridCol: 0},
		{ID: 2, Name: "Hive B", GridRow: 1, GridCol: 2},
	}

	hives, err := svc.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hives) != 2 {
		t.Fatalf("expected 2 hives, got %d", len(hives))
	}
}

func TestListHives_Empty(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	hives, err := svc.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(hives) != 0 {
		t.Errorf("expected empty list, got %d hives", len(hives))
	}
}

func TestListHives_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.List(context.Background(), 1, 99)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestAddHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	apiaryMock.role = "owner"

	hive, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", true, false, false, false, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.Name != "Hive A" {
		t.Errorf("expected name 'Hive A', got %s", hive.Name)
	}
	if !hiveMock.created.Active {
		t.Error("expected hive to be active")
	}
	if !hiveMock.existsByNameCalled {
		t.Fatal("expected ExistsByName to be called")
	}
	if hiveMock.existsByNameApiaryID != 1 || hiveMock.existsByNameName != "Hive A" || hiveMock.existsByNameExcludeID != 0 {
		t.Errorf("unexpected ExistsByName args: apiaryID=%d name=%q excludeID=%d",
			hiveMock.existsByNameApiaryID, hiveMock.existsByNameName, hiveMock.existsByNameExcludeID)
	}
}

func TestAddHive_NeedsFood(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	apiaryMock.role = "owner"

	hive, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", true, false, false, true, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hive.NeedsFood {
		t.Error("expected hive to need food")
	}
	if !hiveMock.created.NeedsFood {
		t.Error("expected created hive to need food")
	}
}

func TestAddHive_MemberCanAdd(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	apiaryMock.role = "member"

	_, err := svc.Add(context.Background(), 2, 1, "Hive B", "langstroth", true, false, false, false, 1, 1)
	if err != nil {
		t.Fatalf("members should be allowed to add hives, got %v", err)
	}
}

func TestAddHive_Inactive(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Add(context.Background(), 1, 1, "Old Hive", "langstroth", false, false, false, false, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hiveMock.created.Active {
		t.Error("expected hive to be inactive")
	}
}

func TestAddHive_NoName(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Add(context.Background(), 1, 1, "", "langstroth", true, false, false, false, 0, 0)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestAddHive_NameTooLong(t *testing.T) {
	svc, _, _ := newTestHiveService()

	name := strings.Repeat("a", 51)
	_, err := svc.Add(context.Background(), 1, 1, name, "langstroth", true, false, false, false, 0, 0)
	if !errors.Is(err, ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong, got %v", err)
	}
}

func TestAddHive_TypeTooLong(t *testing.T) {
	svc, _, _ := newTestHiveService()

	hiveType := strings.Repeat("a", 51)
	_, err := svc.Add(context.Background(), 1, 1, "Hive A", hiveType, true, false, false, false, 0, 0)
	if !errors.Is(err, ErrHiveTypeTooLong) {
		t.Errorf("expected ErrHiveTypeTooLong, got %v", err)
	}
}

func TestAddHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Add(context.Background(), 1, 99, "Hive A", "langstroth", true, false, false, false, 0, 0)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestAddHive_OutOfBounds(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	cases := [][2]int{{-1, 0}, {0, -1}, {3, 0}, {0, 4}}
	for _, c := range cases {
		_, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", true, false, false, false, c[0], c[1])
		if !errors.Is(err, ErrInvalidGridPosition) {
			t.Errorf("row=%d col=%d: expected ErrInvalidGridPosition, got %v", c[0], c[1], err)
		}
	}
}

func TestUpdateHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, Name: "Old", Type: "langstroth", Active: true}

	hive, err := svc.Update(context.Background(), 1, 1, 10, "New Name", "top_bar", false, true, true, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.Name != "New Name" || hive.Type != "top_bar" || hive.Active || !hive.ReadyForHarvest || !hive.Queenless {
		t.Errorf("unexpected hive state: %+v", hive)
	}
	if !hiveMock.existsByNameCalled {
		t.Fatal("expected ExistsByName to be called")
	}
	if hiveMock.existsByNameApiaryID != 1 || hiveMock.existsByNameName != "New Name" || hiveMock.existsByNameExcludeID != 10 {
		t.Errorf("unexpected ExistsByName args: apiaryID=%d name=%q excludeID=%d",
			hiveMock.existsByNameApiaryID, hiveMock.existsByNameName, hiveMock.existsByNameExcludeID)
	}
}

func TestUpdateHive_NeedsFood(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, Name: "Old", Type: "langstroth", Active: true}

	hive, err := svc.Update(context.Background(), 1, 1, 10, "New Name", "top_bar", true, false, false, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hive.NeedsFood {
		t.Errorf("expected hive to need food, got %+v", hive)
	}
}

func TestUpdateHive_NoName(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Update(context.Background(), 1, 1, 10, "", "langstroth", true, false, false, false)
	if !errors.Is(err, ErrNameRequired) {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestUpdateHive_NameTooLong(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	name := strings.Repeat("a", 51)
	_, err := svc.Update(context.Background(), 1, 1, 10, name, "langstroth", true, false, false, false)
	if !errors.Is(err, ErrNameTooLong) {
		t.Errorf("expected ErrNameTooLong, got %v", err)
	}
}

func TestUpdateHive_TypeTooLong(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	hiveType := strings.Repeat("a", 51)
	_, err := svc.Update(context.Background(), 1, 1, 10, "Name", hiveType, true, false, false, false)
	if !errors.Is(err, ErrHiveTypeTooLong) {
		t.Errorf("expected ErrHiveTypeTooLong, got %v", err)
	}
}

func TestUpdateHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Update(context.Background(), 1, 99, 10, "Name", "langstroth", true, false, false, false)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestUpdateHive_DuplicateName(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, Name: "Old", Type: "langstroth", Active: true}
	hiveMock.duplicateName = true

	_, err := svc.Update(context.Background(), 1, 1, 10, "New Name", "top_bar", false, true, true, false)
	if !errors.Is(err, ErrDuplicateHiveName) {
		t.Errorf("expected ErrDuplicateHiveName, got %v", err)
	}
}

func TestUpdateHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Update(context.Background(), 1, 1, 99, "Name", "langstroth", true, false, false, false)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestDeleteHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	if err := svc.Delete(context.Background(), 1, 1, 10); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hiveMock.deletedID != 10 {
		t.Errorf("expected deleted ID 10, got %d", hiveMock.deletedID)
	}
}

func TestDeleteHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	err := svc.Delete(context.Background(), 1, 99, 10)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestDeleteHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	err := svc.Delete(context.Background(), 1, 1, 99)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestGetHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A"}

	hive, err := svc.Get(context.Background(), 1, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.Name != "Hive A" {
		t.Errorf("expected 'Hive A', got %s", hive.Name)
	}
}

func TestGetHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Get(context.Background(), 1, 99, 10)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestGetHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Get(context.Background(), 1, 1, 99)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestMoveHive_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}

	hive, err := svc.Move(context.Background(), 1, 1, 10, 2, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.GridRow != 2 || hive.GridCol != 3 {
		t.Errorf("expected position (2,3), got (%d,%d)", hive.GridRow, hive.GridCol)
	}
}

func TestMoveHive_SamePosition_NoOp(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 1, GridCol: 1}
	hiveMock.occupied = true

	_, err := svc.Move(context.Background(), 1, 1, 10, 1, 1)
	if err != nil {
		t.Fatalf("moving to same position should not error, got %v", err)
	}
}

func TestMoveHive_OutOfBounds(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}

	_, err := svc.Move(context.Background(), 1, 1, 10, 3, 0)
	if !errors.Is(err, ErrInvalidGridPosition) {
		t.Errorf("expected ErrInvalidGridPosition, got %v", err)
	}
}

func TestMoveHive_PositionOccupied(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}
	hiveMock.occupied = true

	_, err := svc.Move(context.Background(), 1, 1, 10, 1, 1)
	if !errors.Is(err, ErrPositionOccupied) {
		t.Errorf("expected ErrPositionOccupied, got %v", err)
	}
}

func TestMoveHive_ApiaryNotFound(t *testing.T) {
	svc, _, _ := newTestHiveService()

	_, err := svc.Move(context.Background(), 1, 99, 10, 0, 0)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestMoveHive_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}

	_, err := svc.Move(context.Background(), 1, 1, 99, 0, 0)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestAddHive_PositionOccupied(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.occupied = true

	_, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", true, false, false, false, 0, 0)
	if !errors.Is(err, ErrPositionOccupied) {
		t.Errorf("expected ErrPositionOccupied, got %v", err)
	}
}

func TestAddHive_DuplicateName(t *testing.T) {
	svc, apiaryMock, hiveMock := newTestHiveService()
	apiaryMock.apiary = &model.Apiary{ID: 1, GridRows: 3, GridCols: 4}
	hiveMock.duplicateName = true

	_, err := svc.Add(context.Background(), 1, 1, "Hive A", "langstroth", true, false, false, false, 0, 0)
	if !errors.Is(err, ErrDuplicateHiveName) {
		t.Errorf("expected ErrDuplicateHiveName, got %v", err)
	}
}

func newChangeApiaryService() (*HiveService, *mockMultiApiaryReader, *mockHiveRepo) {
	apiaryMock := &mockMultiApiaryReader{
		apiaries: make(map[int64]*model.Apiary),
		roles:    make(map[int64]string),
	}
	hiveMock := &mockHiveRepo{}
	svc := NewHiveService(apiaryMock, hiveMock)
	return svc, apiaryMock, hiveMock
}

func TestChangeApiary_Success(t *testing.T) {
	svc, apiaryMock, hiveMock := newChangeApiaryService()
	apiaryMock.apiaries[1] = &model.Apiary{ID: 1, GridRows: 2, GridCols: 2}
	apiaryMock.apiaries[2] = &model.Apiary{ID: 2, GridRows: 2, GridCols: 2}
	apiaryMock.roles[1] = "owner"
	apiaryMock.roles[2] = "owner"
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, GridRow: 0, GridCol: 0}
	hiveMock.hivesByID = map[int64][]*model.Hive{
		2: {{ID: 5, ApiaryID: 2, GridRow: 0, GridCol: 0}},
	}

	hive, err := svc.ChangeApiary(context.Background(), 1, 1, 10, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hive.ApiaryID != 2 {
		t.Errorf("expected ApiaryID=2, got %d", hive.ApiaryID)
	}
	if hive.GridRow != 0 || hive.GridCol != 1 {
		t.Errorf("expected position (0,1), got (%d,%d)", hive.GridRow, hive.GridCol)
	}
}

func TestChangeApiary_SameApiary(t *testing.T) {
	svc, _, _ := newChangeApiaryService()

	_, err := svc.ChangeApiary(context.Background(), 1, 1, 10, 1)
	if !errors.Is(err, ErrSameApiary) {
		t.Errorf("expected ErrSameApiary, got %v", err)
	}
}

func TestChangeApiary_SourceNotFound(t *testing.T) {
	svc, _, _ := newChangeApiaryService()

	_, err := svc.ChangeApiary(context.Background(), 1, 99, 10, 2)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestChangeApiary_HiveNotFound(t *testing.T) {
	svc, apiaryMock, _ := newChangeApiaryService()
	apiaryMock.apiaries[1] = &model.Apiary{ID: 1, GridRows: 2, GridCols: 2}
	apiaryMock.roles[1] = "owner"

	_, err := svc.ChangeApiary(context.Background(), 1, 1, 99, 2)
	if !errors.Is(err, ErrHiveNotFound) {
		t.Errorf("expected ErrHiveNotFound, got %v", err)
	}
}

func TestChangeApiary_TargetNotFound(t *testing.T) {
	svc, apiaryMock, hiveMock := newChangeApiaryService()
	apiaryMock.apiaries[1] = &model.Apiary{ID: 1, GridRows: 2, GridCols: 2}
	apiaryMock.roles[1] = "owner"
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}

	_, err := svc.ChangeApiary(context.Background(), 1, 1, 10, 99)
	if !errors.Is(err, ErrApiaryNotFound) {
		t.Errorf("expected ErrApiaryNotFound, got %v", err)
	}
}

func TestChangeApiary_TargetFull(t *testing.T) {
	svc, apiaryMock, hiveMock := newChangeApiaryService()
	apiaryMock.apiaries[1] = &model.Apiary{ID: 1, GridRows: 2, GridCols: 2}
	apiaryMock.apiaries[2] = &model.Apiary{ID: 2, GridRows: 1, GridCols: 1}
	apiaryMock.roles[1] = "owner"
	apiaryMock.roles[2] = "member"
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1}
	hiveMock.hivesByID = map[int64][]*model.Hive{
		2: {{ID: 5, ApiaryID: 2, GridRow: 0, GridCol: 0}},
	}

	_, err := svc.ChangeApiary(context.Background(), 1, 1, 10, 2)
	if !errors.Is(err, ErrTargetApiaryFull) {
		t.Errorf("expected ErrTargetApiaryFull, got %v", err)
	}
}

func TestChangeApiary_DuplicateName(t *testing.T) {
	svc, apiaryMock, hiveMock := newChangeApiaryService()
	apiaryMock.apiaries[1] = &model.Apiary{ID: 1, GridRows: 2, GridCols: 2}
	apiaryMock.apiaries[2] = &model.Apiary{ID: 2, GridRows: 2, GridCols: 2}
	apiaryMock.roles[1] = "owner"
	apiaryMock.roles[2] = "owner"
	hiveMock.hive = &model.Hive{ID: 10, ApiaryID: 1, Name: "Hive A", GridRow: 0, GridCol: 0}
	hiveMock.duplicateName = true

	_, err := svc.ChangeApiary(context.Background(), 1, 1, 10, 2)
	if !errors.Is(err, ErrDuplicateHiveName) {
		t.Errorf("expected ErrDuplicateHiveName, got %v", err)
	}
}
