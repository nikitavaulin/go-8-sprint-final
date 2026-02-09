package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcel.Number, err = store.Add(parcel)
	assert.NoError(t, err)
	assert.NotEmpty(t, parcel.Number)

	// get
	parcelCopy, err := store.Get(parcel.Number)
	assert.NoError(t, err)
	assert.Equal(t, parcel.Number, parcelCopy.Number)
	assert.Equal(t, parcel.Address, parcelCopy.Address)
	assert.Equal(t, parcel.Client, parcelCopy.Client)
	assert.Equal(t, parcel.CreatedAt, parcelCopy.CreatedAt)

	// delete
	err = store.Delete(parcelCopy.Number)
	assert.NoError(t, err)
	_, err = store.Get(parcelCopy.Number)
	assert.ErrorIs(t, err, ErrorParcelNotFound)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	parcelCopy, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, parcelCopy.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set status
	err = store.SetStatus(parcel.Number, ParcelStatusRegistered)
	require.NoError(t, err)

	// check
	status, err := store.GetStatus(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusRegistered, status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		parcelOriginal, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, parcelOriginal.Number, parcel.Number)
		assert.Equal(t, parcelOriginal.Address, parcel.Address)
		assert.Equal(t, parcelOriginal.Client, parcel.Client)
		assert.Equal(t, parcelOriginal.CreatedAt, parcel.CreatedAt)
		assert.Equal(t, parcelOriginal.Status, parcel.Status)
	}
}
