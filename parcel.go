package main

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrorParcelNotFound = errors.New("parcel not found")
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec(
		"INSERT INTO parcel (client, status, address, created_at) VALUES (:client, :status, :address, :date)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("date", p.CreatedAt))
	if err != nil {
		return 0, fmt.Errorf("add new parcel error: %w", err)
	}

	parcelId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted parcel id error: %w", err)
	}

	p.Number = int(parcelId)

	return p.Number, nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow("SELECT client, status, address, created_at FROM parcel WHERE number = :id", sql.Named("id", number))
	err := row.Scan(&p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Parcel{}, ErrorParcelNotFound
		}
		return p, fmt.Errorf("get parcel by number error: %w", err)
	}

	p.Number = number

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, status, address, created_at FROM parcel WHERE client = :id", sql.Named("id", client))
	if err != nil {
		return nil, fmt.Errorf("get parcels by client error: %w", err)
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p = Parcel{}
		err = rows.Scan(&p.Number, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan row attributes error: %w", err)
		}
		p.Client = client

		parcels = append(parcels, p)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iter rows error: %w", rows.Err())
	}

	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		"UPDATE parcel SET status = :status WHERE number = :id",
		sql.Named("status", status),
		sql.Named("id", number))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrorParcelNotFound
		}
		return fmt.Errorf("status update error: %w", err)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec(
		"UPDATE parcel SET address = :address WHERE number = :id AND status = :status",
		sql.Named("address", address),
		sql.Named("status", ParcelStatusRegistered),
		sql.Named("id", number))
	if err != nil {
		return fmt.Errorf("SetAddress: address update error: %w", err)
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	_, err := s.db.Exec(
		"DELETE FROM parcel WHERE number = :id AND status = :status",
		sql.Named("id", number),
		sql.Named("status", ParcelStatusRegistered))
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	return nil
}
