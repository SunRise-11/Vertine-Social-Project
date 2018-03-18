package repo

import (
	"fmt"

	"github.com/Coderockr/vitrine-social/server/security"

	"github.com/Coderockr/vitrine-social/server/model"
	"github.com/jmoiron/sqlx"
)

// OrganizationRepository is a implementation for Postgres
type OrganizationRepository struct {
	db       *sqlx.DB
	catRepo  *CategoryRepository
	needRepo *NeedRepository
}

// NewOrganizationRepository creates a new repository
func NewOrganizationRepository(db *sqlx.DB) *OrganizationRepository {
	r := &OrganizationRepository{
		db:      db,
		catRepo: NewCategoryRepository(db),
	}
	r.needRepo = NewNeedRepository(db, r)

	return r
}

// Get one Organization from database
func (r *OrganizationRepository) Get(id int64) (*model.Organization, error) {
	o := &model.Organization{}
	err := r.db.Get(o, "SELECT * FROM organizations WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&o.Images, "SELECT * FROM organizations_images WHERE organization_id = $1", id)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&o.Needs, "SELECT * FROM needs WHERE organization_id = $1", id)
	if err != nil {
		return nil, err
	}

	for i := range o.Needs {
		o.Needs[i].Category, err = r.catRepo.Get(o.Needs[i].CategoryID)
		if err != nil {
			fmt.Println("test?")
			return nil, err
		}

		o.Needs[i].Images, err = r.needRepo.getNeedImages(&o.Needs[i])
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

// Create receives a Organization and creates it in the database, returning the updated Organization or error if failed
func (r *OrganizationRepository) Create(o model.Organization) (model.Organization, error) {
	row := r.db.QueryRow(
		`INSERT INTO organizations (name, logo, address, phone, resume, video, email, slug, password)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`,
		o.Name,
		o.Logo,
		o.Address,
		o.Phone,
		o.Resume,
		o.Video,
		o.Email,
		o.Slug,
		o.Password,
	)

	err := row.Scan(&o.ID)

	if err != nil {
		return o, err
	}

	return o, nil
}

// GetByEmail returns a organization by its email
func (r *OrganizationRepository) GetByEmail(email string) (*model.Organization, error) {
	o := model.Organization{}
	err := r.db.Get(&o, `SELECT * FROM organizations WHERE email = $1`, email)
	return &o, err
}

// GetUserByEmail returns a organization user by its email
func (r *OrganizationRepository) GetUserByEmail(email string) (model.User, error) {
	o, err := r.GetByEmail(email)
	if err != nil {
		return model.User{}, err
	}
	return o.User, nil
}

// ResetPasswordTo resets the organization password to the value informed
func (r *OrganizationRepository) ResetPasswordTo(o *model.Organization, password string) error {
	hash, err := security.GenerateHash(password)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`UPDATE organizations SET password = $1 WHERE id = $2`, hash, o.ID)
	if err != nil {
		return err
	}
	o.Password = hash
	return nil
}
