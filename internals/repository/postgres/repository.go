package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"practice/internals/entity"
	"practice/internals/params/errs"
)

type UserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (repo *UserRepository) userTypeAssertion(data map[string]interface{}) (*entity.User, error) {
	id, ok := data["id"].(string)
	if !ok {
		return nil, fmt.Errorf("unvalid id")
	}
	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("unvalid name")
	}
	email, ok := data["email"].(string)
	if !ok {
		return nil, fmt.Errorf("unvalid email")
	}
	phone, ok := data["phone_number"].(string)
	if !ok {
		return nil, fmt.Errorf("unvalid phone_number")
	}
	user := &entity.User{
		ID:          id,
		Name:        name,
		Email:       email,
		PhoneNumber: phone,
	}
	return user, nil

}

func (repo *UserRepository) mapToUser(data map[string]interface{}) (*entity.User, error) {
	user, err := repo.userTypeAssertion(data)
	if err != nil {

		return nil, err
	}
	rawAddresses, ok := data["addresses"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unvalid addresses")
	}

	for _, raw := range rawAddresses {
		addr, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unvalid addresses")
		}
		street, _ := addr["street"].(string)
		city, _ := addr["city"].(string)
		state, _ := addr["state"].(string)
		zipCode, _ := addr["zip_code"].(string)
		country, _ := addr["country"].(string)

		user.Addresses = append(user.Addresses, entity.Address{
			Street:  street,
			City:    city,
			State:   state,
			ZipCode: zipCode,
			Country: country,
		})
	}

	return user, nil
}

func (repo *UserRepository) Insert(data map[string]interface{}) error {
	logger := repo.logger.With(
		slog.String("component", "UserRepository"),
		slog.String("method", "Insert"),
	)

	u, err := repo.mapToUser(data)
	if err != nil {
		logger.Error("user mapping failed", slog.Any("error", err))
		return fmt.Errorf("mapping faild: %w", err)
	}

	tx, err := repo.db.Begin()
	if err != nil {
		logger.Error("transaction begin failed",
			slog.String("user_id", u.ID),
			slog.Any("error", err),
		)
		return fmt.Errorf("transaction begin failed: %w", err)
	}
	defer tx.Rollback()

	const userQ = `
		INSERT INTO users (id, name, email, phone_number)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO NOTHING
	`
	if _, err := tx.Exec(userQ, u.ID, u.Name, u.Email, u.PhoneNumber); err != nil {
		logger.Error("insert failed",
			slog.String("user_id", u.ID),
			slog.Any("error", err),
		)
		return fmt.Errorf("insert user failed: %w", err)
	}

	const addrQ = `
		INSERT INTO addresses (user_id, street, city, state, zip_code, country)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, addr := range u.Addresses {
		if _, err := tx.Exec(addrQ,
			u.ID, addr.Street, addr.City, addr.State, addr.ZipCode, addr.Country,
		); err != nil {
			return fmt.Errorf("insert address failed: %w", err)
		}
	}
	logger.Info("user inserted successfully",
		slog.String("user_id", u.ID),
	)

	return tx.Commit()
}

func (repo *UserRepository) GetByID(id string) (*entity.User, error) {
	logger := repo.logger.With(
		slog.String("component", "UserRepository"),
		slog.String("method", "GetByID"),
		slog.String("user_id", id),
	)

	const q = `
		SELECT
			u.id,
			u.name,
			u.email,
			u.phone_number,
			a.street,
			a.city,
			a.state,
			a.zip_code,
			a.country
		FROM users u
		LEFT JOIN addresses a ON a.user_id = u.id
		WHERE u.id = $1
	`

	rows, err := repo.db.Query(q, id)
	if err != nil {
		logger.Error("query failed", slog.Any("error", err))
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var user *entity.User

	for rows.Next() {
		var street, city, state, zipCode, country sql.NullString

		if user == nil {
			user = &entity.User{}
		}

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PhoneNumber,
			&street,
			&city,
			&state,
			&zipCode,
			&country,
		); err != nil {
			logger.Error("scan failed", slog.Any("error", err))
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		if street.Valid {
			user.Addresses = append(user.Addresses, entity.Address{
				Street:  street.String,
				City:    city.String,
				State:   state.String,
				ZipCode: zipCode.String,
				Country: country.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		logger.Error("rows failed", slog.Any("error", err))
		return nil, fmt.Errorf("failed reading rows: %w", err)
	}

	if user == nil {
		logger.Info("user not found")
		return nil, errs.ErrUserNotFound
	}

	logger.Info("user found")

	return user, nil
}
