package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/layerhub"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ExtContext interface {
	sqlx.ExecerContext
	sqlx.QueryerContext
}

type Config struct {
	DSN             string
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

type MySQLDB struct {
	db *sqlx.DB
}

func New(conf *Config) (layerhub.DB, error) {
	db, err := sqlx.Open("mysql", conf.DSN)
	if err != nil {
		return nil, err
	}

	if conf.ConnMaxIdleTime != 0 {
		db.SetConnMaxIdleTime(conf.ConnMaxIdleTime)
	}
	if conf.ConnMaxLifetime != 0 {
		db.SetConnMaxLifetime(conf.ConnMaxLifetime)
	}
	if conf.MaxIdleConns != 0 {
		db.SetMaxIdleConns(conf.MaxIdleConns)
	}
	if conf.MaxOpenConns != 0 {
		db.SetMaxOpenConns(conf.MaxOpenConns)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &MySQLDB{db}, nil
}

func (s *MySQLDB) PutUser(ctx context.Context, user *layerhub.User) error {
	query := `INSERT INTO users (
        id,
        first_name,
        last_name,
        email,
        phone,
        avatar,
        email_verified,
        phone_verified,
        kind,
        password_hash,
        source,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        first_name=VALUES(first_name),
        last_name=VALUES(last_name),
        email=VALUES(email),
        phone=VALUES(phone),
        avatar=VALUES(avatar),
        email_verified=VALUES(email_verified),
        phone_verified=VALUES(phone_verified),
        kind=VALUES(kind),
        password_hash=VALUES(password_hash),
        updated_at=VALUES(updated_at)
    `

	_, err := s.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Phone,
		user.Avatar,
		user.EmailVerified,
		user.PhoneVerified,
		user.Kind,
		user.PasswordHash,
		user.Source,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindUsers(ctx context.Context, filter *layerhub.Filter) ([]layerhub.User, error) {
	query := `SELECT * FROM users `

	where, args := filterToQuery("users", filter)
	users := []layerhub.User{}

	err := s.db.SelectContext(ctx, &users, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return users, nil
}

func (s *MySQLDB) BatchCreateFonts(ctx context.Context, fonts []layerhub.Font) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	defer tx.Rollback()

	batchSize := 500

	for batchStart := 0; batchStart < len(fonts); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd >= len(fonts) {
			batchEnd = len(fonts)
		}

		query := `INSERT INTO fonts (
            id,
            family,
            full_name,
            postscript_name,
            preview,
            style,
            url,
            user_id,
            category
        ) VALUES `

		args := []any{}
		values := []string{}
		for _, pf := range fonts[batchStart:batchEnd] {
			values = append(values, "(?, ?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(
				args,
				pf.ID,
				pf.Family,
				pf.FullName,
				pf.PostscriptName,
				pf.Preview,
				pf.Style,
				pf.URL,
				pf.UserID,
				pf.Category,
			)
		}
		query += strings.Join(values, ",")

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return errors.E(errors.KindUnexpected, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutCustomer(ctx context.Context, company *layerhub.Customer) error {
	query := `INSERT INTO customers (
        id,
        first_name,
        last_name,
        email,
        company_id,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        first_name=VALUES(first_name),
        last_name=VALUES(last_name),
        email=VALUES(email),
        updated_at=VALUES(updated_at)
    `

	_, err := s.db.ExecContext(
		ctx,
		query,
		company.ID,
		company.FirstName,
		company.LastName,
		company.Email,
		company.CompanyID,
		company.CreatedAt,
		company.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindCustomers(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Customer, error) {
	query := `SELECT * FROM customers `
	where, args := filterToQuery("customers", filter)
	customers := []layerhub.Customer{}

	err := s.db.SelectContext(ctx, &customers, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return customers, nil
}

func (s *MySQLDB) CountCustomers(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM customers `
	where, args := filterToQuery("customers", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteCustomer(ctx context.Context, id string) error {
	query := `DELETE FROM customers WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutCompany(ctx context.Context, company *layerhub.Company) error {
	query := `INSERT INTO companies (
        id,
        name,
        user_id,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        name=VALUES(name),
        updated_at=VALUES(updated_at)
    `

	_, err := s.db.ExecContext(
		ctx,
		query,
		company.ID,
		company.Name,
		company.UserID,
		company.CreatedAt,
		company.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindCompanies(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Company, error) {
	query := `SELECT * FROM companies `
	where, args := filterToQuery("companies", filter)
	companies := []layerhub.Company{}

	err := s.db.SelectContext(ctx, &companies, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return companies, nil
}

func (s *MySQLDB) CountCompanies(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM companies `
	where, args := filterToQuery("companies", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteCompany(ctx context.Context, id string) error {
	query := `DELETE FROM companies WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindFonts(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Font, error) {
	query := `SELECT fonts.* FROM fonts `
	where, args := filterToQuery("fonts", filter)
	fonts := []layerhub.Font{}

	if filter != nil && filter.FontEnabled != nil {
		if *filter.FontEnabled {
			query += "INNER JOIN enabled_fonts ON fonts.id = enabled_fonts.font_id AND enabled_fonts.user_id = ? "
		} else {
			query += "LEFT JOIN enabled_fonts ON fonts.id = enabled_fonts.font_id AND enabled_fonts.user_id = ? "
		}
		args = append([]any{filter.UserID}, args...)
	}

	err := s.db.SelectContext(ctx, &fonts, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return fonts, nil
}

func (s *MySQLDB) CountFonts(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM fonts `
	where, args := filterToQuery("fonts", filter)
	count := []CountRow{}

	if filter != nil && filter.FontEnabled != nil {
		if *filter.FontEnabled {
			query += "INNER JOIN enabled_fonts ON fonts.id = enabled_fonts.font_id AND enabled_fonts.user_id = ? "
		} else {
			query += "LEFT JOIN enabled_fonts ON fonts.id = enabled_fonts.font_id AND enabled_fonts.user_id = ? "
		}
		args = append([]any{filter.UserID}, args...)
	}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s MySQLDB) PutFont(ctx context.Context, font *layerhub.Font) error {
	query := `INSERT INTO fonts (
        id,
        full_name,
        family,
        postscript_name,
        preview,
        style,
        url,
        category,
        user_id
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        full_name=VALUES(full_name),
        family=VALUES(family),
        postscript_name=VALUES(postscript_name),
        preview=VALUES(preview),
        style=VALUES(style),
        url=VALUES(url),
        category=VALUES(category),
        user_id=VALUES(user_id)
    `
	_, err := s.db.ExecContext(
		ctx,
		query,
		font.ID,
		font.FullName,
		font.Family,
		font.PostscriptName,
		font.Preview,
		font.Style,
		font.URL,
		font.Category,
		font.UserID,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s MySQLDB) DeleteFont(ctx context.Context, id string) error {
	query := `DELETE FROM fonts WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutTemplate(ctx context.Context, template *layerhub.Template) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	defer tx.Rollback()

	query := `INSERT INTO templates (
        id,
        short_id,
        name,
        type,
        published,
        preview,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        short_id=VALUES(short_id),
        name=VALUES(name),
        type=VALUES(type),
        published=VALUES(published),
        preview=VALUES(preview),
        updated_at=VALUES(updated_at)
    `

	_, err = tx.ExecContext(
		ctx,
		query,
		template.ID,
		template.ShortID,
		template.Name,
		template.Type,
		template.Published,
		template.Preview,
		template.CreatedAt,
		template.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.putFrame(ctx, tx, &layerhub.Frame{
		ID:         template.ID,
		Name:       template.Frame.Name,
		Width:      template.Frame.Width,
		Height:     template.Frame.Height,
		Visibility: layerhub.FramePrivate,
	})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.putTemplateTags(ctx, tx, template)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.putTemplateColors(ctx, tx, template)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.putTemplateMetadata(ctx, tx, template)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindTemplates(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Template, error) {
	query := `SELECT * FROM templates `
	where, args := filterToQuery("templates", filter)
	templates := []layerhub.Template{}

	err := s.db.SelectContext(ctx, &templates, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	for i, template := range templates {
		frames, err := s.FindFrames(ctx, &layerhub.Filter{ID: template.ID})
		if err != nil {
			return nil, err
		}
		tags, err := s.getTemplateTags(ctx, template.ID)
		if err != nil {
			return nil, err
		}
		colors, err := s.getTemplateColors(ctx, template.ID)
		if err != nil {
			return nil, err
		}
		metadata, err := s.getTemplateMetadata(ctx, template.ID)
		if err != nil {
			return nil, err
		}
		if len(frames) > 0 {
			templates[i].Frame = frames[0]
		}
		templates[i].Tags = tags
		templates[i].Colors = colors
		templates[i].Metadata = metadata
	}

	return templates, nil
}

func (s *MySQLDB) CountTemplates(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM templates `
	where, args := filterToQuery("templates", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteTemplate(ctx context.Context, id string) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	defer tx.Rollback()

	query := `DELETE FROM templates WHERE id = ?`

	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.deleteTemplateColors(ctx, tx, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.deleteTemplateTags(ctx, tx, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.deleteFrame(ctx, tx, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutFrame(ctx context.Context, frame *layerhub.Frame) error {
	return s.putFrame(ctx, s.db, frame)
}

func (s *MySQLDB) putFrame(ctx context.Context, ext ExtContext, frame *layerhub.Frame) error {
	query := `INSERT INTO frames (
        id,
        name,
        visibility,
        width,
        height,
        unit,
        preview
    ) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        name=VALUES(name),
        visibility=VALUES(visibility),
        width=VALUES(width),
        height=VALUES(height),
        unit=VALUES(unit),
        preview=VALUES(preview)
    `

	_, err := ext.ExecContext(
		ctx,
		query,
		frame.ID,
		frame.Name,
		frame.Visibility,
		frame.Width,
		frame.Height,
		frame.Unit,
		frame.Preview,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindFrames(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Frame, error) {
	query := `SELECT * FROM frames `
	where, args := filterToQuery("frames", filter)
	frames := []layerhub.Frame{}

	err := s.db.SelectContext(ctx, &frames, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return frames, nil
}

func (s *MySQLDB) CountFrames(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM frames `
	where, args := filterToQuery("frames", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteFrame(ctx context.Context, id string) error {
	return s.deleteFrame(ctx, s.db, id)
}

func (s *MySQLDB) deleteFrame(ctx context.Context, ext ExtContext, id string) error {
	query := `DELETE FROM frames WHERE id = ?`

	_, err := ext.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutProject(ctx context.Context, project *layerhub.Project) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	defer tx.Rollback()

	query := `INSERT INTO projects (
        id,
        short_id,
        name,
        type,
        preview,
        user_id,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        short_id=VALUES(short_id),
        name=VALUES(name),
        type=VALUES(type),
        preview=VALUES(preview),
        updated_at=VALUES(updated_at)
    `

	_, err = tx.ExecContext(
		ctx,
		query,
		project.ID,
		project.ShortID,
		project.Name,
		project.Type,
		project.Preview,
		project.UserID,
		project.CreatedAt,
		project.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.putFrame(ctx, tx, &layerhub.Frame{
		ID:         project.ID,
		Name:       project.Frame.Name,
		Width:      project.Frame.Width,
		Height:     project.Frame.Height,
		Visibility: layerhub.FramePrivate,
	})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindProjects(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Project, error) {
	query := `SELECT * FROM projects `
	where, args := filterToQuery("projects", filter)
	projects := []layerhub.Project{}

	err := s.db.SelectContext(ctx, &projects, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	for i, p := range projects {
		frames, err := s.FindFrames(ctx, &layerhub.Filter{ID: p.ID})
		if err != nil {
			return nil, err
		}
		if len(frames) == 0 {
			continue
		}

		projects[i].Frame = frames[0]
	}

	return projects, nil
}

func (s *MySQLDB) CountProjects(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM projects `
	where, args := filterToQuery("projects", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteProject(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutComponent(ctx context.Context, component *layerhub.Component) error {
	query := `INSERT INTO components (
        id,
        name,
        preview,
        user_id,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        name=VALUES(name),
        preview=VALUES(preview),
        updated_at=VALUES(updated_at)
    `

	_, err := s.db.ExecContext(
		ctx,
		query,
		component.ID,
		component.Name,
		component.Preview,
		component.UserID,
		component.CreatedAt,
		component.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindComponents(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Component, error) {
	query := `SELECT * FROM components `
	where, args := filterToQuery("components", filter)
	components := []layerhub.Component{}

	err := s.db.SelectContext(ctx, &components, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return components, nil
}

func (s *MySQLDB) CountComponents(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM components `
	where, args := filterToQuery("components", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteComponent(ctx context.Context, id string) error {
	query := `DELETE FROM components WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) PutUpload(ctx context.Context, upload *layerhub.Upload) error {
	query := `INSERT INTO uploads (
        id,
        name,
        content_type,
        folder,
        type,
        url,
        user_id,
        created_at,
        updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        name=VALUES(name),
        content_type=VALUES(content_type),
        folder=VALUES(folder),
        type=VALUES(type),
        url=VALUES(url),
        updated_at=VALUES(updated_at)
    `

	_, err := s.db.ExecContext(
		ctx,
		query,
		upload.ID,
		upload.Name,
		upload.ContentType,
		upload.Folder,
		upload.Type,
		upload.URL,
		upload.UserID,
		upload.CreatedAt,
		upload.UpdatedAt,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindUploads(ctx context.Context, filter *layerhub.Filter) ([]layerhub.Upload, error) {
	query := `SELECT * FROM uploads `
	where, args := filterToQuery("uploads", filter)
	uploads := []layerhub.Upload{}

	err := s.db.SelectContext(ctx, &uploads, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return uploads, nil
}

func (s *MySQLDB) CountUploads(ctx context.Context, filter *layerhub.Filter) (int, error) {
	query := `SELECT COUNT(*) AS count FROM uploads `
	where, args := filterToQuery("uploads", filter)
	count := []CountRow{}

	err := s.db.SelectContext(ctx, &count, query+where, args...)
	if err != nil {
		return 0, errors.E(errors.KindUnexpected, err)
	}

	return count[0].Count, nil
}

func (s *MySQLDB) DeleteUpload(ctx context.Context, id string) error {
	query := `DELETE FROM uploads WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindEnabledFonts(ctx context.Context, userID string) ([]layerhub.EnabledFont, error) {
	query := `SELECT * FROM enabled_fonts `
	where, args := filterToQuery("enabled_fonts", &layerhub.Filter{
		UserID: userID,
	})
	fonts := []layerhub.EnabledFont{}

	err := s.db.SelectContext(ctx, &fonts, query+where, args...)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	return fonts, nil
}

func (s *MySQLDB) BatchCreateEnabledFonts(ctx context.Context, fonts []*layerhub.EnabledFont) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	defer tx.Rollback()

	batchSize := 500

	for batchStart := 0; batchStart < len(fonts); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd >= len(fonts) {
			batchEnd = len(fonts)
		}

		query := `INSERT INTO enabled_fonts (
            id,
            user_id,
            font_id
        ) VALUES `

		args := []any{}
		values := []string{}
		for _, f := range fonts[batchStart:batchEnd] {
			values = append(values, "(?, ?, ?)")
			args = append(
				args,
				f.ID,
				f.UserID,
				f.FontID,
			)
		}
		query += strings.Join(values, ",")

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return errors.E(errors.KindUnexpected, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) BatchDeleteEnabledFonts(ctx context.Context, ids []string) error {
	args := make([]string, len(ids))
	values := make([]any, len(ids))
	for i, id := range ids {
		args[i] = "?"
		values[i] = id
	}

	if len(ids) != 0 {
		query := fmt.Sprintf("DELETE FROM enabled_fonts WHERE id IN (%s)", strings.Join(args, ","))

		_, err := s.db.ExecContext(ctx, query, values...)
		if err != nil {
			return errors.E(errors.KindUnexpected, err)
		}
	}

	return nil
}

func (s *MySQLDB) PutSubscriptionPlan(ctx context.Context, plan *layerhub.SubscriptionPlan) error {
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}
	defer tx.Rollback()

	query := `INSERT INTO subscription_plans (
        id,
        provider,
        external_id,
        name,
        description,
        external_product_id,
        auto_bill_outstanding,
        setup_fee,
        max_templates
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE 
        name=VALUES(name),
        description=VALUES(description),
        external_product_id=VALUES(external_product_id),
        auto_bill_outstanding=VALUES(auto_bill_outstanding),
        setup_fee=VALUES(setup_fee),
        max_templates=VALUES(max_templates)
    `

	_, err = tx.ExecContext(
		ctx,
		query,
		plan.ID,
		plan.Provider,
		plan.ExternalID,
		plan.Name,
		plan.Description,
		plan.ExternalProductID,
		plan.AutoBillOutstanding,
		plan.SetupFee,
		plan.MaxTemplates,
	)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = s.putSubscriptionPlanBillings(ctx, tx, plan)
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	err = tx.Commit()
	if err != nil {
		return errors.E(errors.KindUnexpected, err)
	}

	return nil
}

func (s *MySQLDB) FindSubscriptionPlans(ctx context.Context) ([]layerhub.SubscriptionPlan, error) {
	query := `SELECT * FROM subscription_plans`
	plans := []layerhub.SubscriptionPlan{}

	err := s.db.SelectContext(ctx, &plans, query)
	if err != nil {
		return nil, errors.E(errors.KindUnexpected, err)
	}

	for i, plan := range plans {
		billings, err := s.getSubscriptionPlanBillings(ctx, plan.ID)
		if err != nil {
			return nil, errors.E(errors.KindUnexpected, err)
		}
		plans[i].Billing = billings
	}

	return plans, nil
}

func (s *MySQLDB) deleteTemplateTags(ctx context.Context, ext ExtContext, templateID string) error {
	delQuery := `DELETE FROM template_tags WHERE template_id = ?`
	_, err := ext.ExecContext(ctx, delQuery, templateID)
	return err
}

func (s *MySQLDB) deleteTemplateColors(ctx context.Context, ext ExtContext, templateID string) error {
	delQuery := `DELETE FROM template_colors WHERE template_id = ?`
	_, err := ext.ExecContext(ctx, delQuery, templateID)
	return err
}

func (s *MySQLDB) putTemplateMetadata(ctx context.Context, ext ExtContext, template *layerhub.Template) error {
	query := `INSERT INTO template_metadata (
        id,
        license,
        orientation
    ) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE
        license=VALUES(license),
        orientation=VALUES(orientation)
    `

	_, err := ext.ExecContext(
		ctx,
		query,
		template.ID,
		template.Metadata.License,
		template.Metadata.Orientation,
	)
	return err
}

func (s *MySQLDB) putTemplateTags(ctx context.Context, ext ExtContext, template *layerhub.Template) error {
	err := s.deleteTemplateTags(ctx, ext, template.ID)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO template_tags (
        id,
        template_id,
        tag,
        position
    ) VALUES (?, ?, ?, ?)`
	for pos, t := range template.Tags {
		_, err := ext.ExecContext(ctx, insertQuery, layerhub.UniqueID("tag"), template.ID, t, pos)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySQLDB) putTemplateColors(ctx context.Context, ext ExtContext, template *layerhub.Template) error {
	err := s.deleteTemplateColors(ctx, ext, template.ID)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO template_colors (
        id,
        template_id,
        color,
        position
    ) VALUES (?, ?, ?, ?)`
	for pos, c := range template.Colors {
		_, err := ext.ExecContext(ctx, insertQuery, layerhub.UniqueID("color"), template.ID, c, pos)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySQLDB) getTemplateMetadata(ctx context.Context, templateID string) (layerhub.Metadata, error) {
	var metadata layerhub.Metadata
	query := `SELECT * FROM template_metadata WHERE id = ?`
	err := s.db.GetContext(ctx, &metadata, query, templateID)
	return metadata, err
}

func (s *MySQLDB) getTemplateTags(ctx context.Context, templateID string) ([]string, error) {
	query := `SELECT * FROM template_tags WHERE template_id = ? ORDER BY position`
	rows, err := s.db.QueryxContext(ctx, query, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []string{}
	for rows.Next() {
		row := struct {
			ID         string `db:"id"`
			TemplateID string `db:"template_id"`
			Tag        string `db:"tag"`
			Position   int    `db:"position"`
		}{}
		err := rows.StructScan(&row)
		if err != nil {
			return nil, err
		}
		tags = append(tags, row.Tag)
	}
	return tags, nil
}

func (s *MySQLDB) getTemplateColors(ctx context.Context, templateID string) ([]string, error) {
	query := `SELECT * FROM template_colors WHERE template_id = ? ORDER BY position`
	rows, err := s.db.QueryxContext(ctx, query, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colors := []string{}
	for rows.Next() {
		row := struct {
			ID         string `db:"id"`
			TemplateID string `db:"template_id"`
			Color      string `db:"color"`
			Position   int    `db:"position"`
		}{}
		err := rows.StructScan(&row)
		if err != nil {
			return nil, err
		}
		colors = append(colors, row.Color)
	}
	return colors, nil
}

func (s *MySQLDB) getSubscriptionPlanBillings(ctx context.Context, planID string) ([]*layerhub.Billing, error) {
	query := `SELECT * FROM subscription_plan_billings WHERE subscription_plan_id = ?`
	billings := []*layerhub.Billing{}

	err := s.db.SelectContext(ctx, &billings, query, planID)
	if err != nil {
		return nil, err
	}

	return billings, nil
}

func (s *MySQLDB) putSubscriptionPlanBillings(ctx context.Context, ext ExtContext, plan *layerhub.SubscriptionPlan) error {
	delQuery := `DELETE FROM subscription_plan_billings WHERE subscription_plan_id = ?`

	_, err := ext.ExecContext(
		ctx,
		delQuery,
		plan.ID,
	)
	if err != nil {
		return err
	}

	insertQuery := fmt.Sprintf(`INSERT INTO subscription_plan_billings (
        id,
        %s,
        price,
        subscription_plan_id
    ) VALUES (?, ?, ?, ?)`, "`interval`")

	for _, b := range plan.Billing {
		_, err := ext.ExecContext(
			ctx,
			insertQuery,
			b.ID,
			b.Interval,
			b.Price,
			plan.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func filterToQuery(table string, filter *layerhub.Filter) (string, []any) {
	query := ""
	args := []any{}
	conds := []string{}

	if filter != nil {
		if filter.Email != "" {
			conds = append(conds, fmt.Sprintf("%s.email = ?", table))
			args = append(args, filter.Email)
		}
		if filter.PostscriptName != "" {
			conds = append(conds, fmt.Sprintf("%s.postscript_name = ?", table))
			args = append(args, filter.PostscriptName)
		}
		if filter.ID != "" {
			conds = append(conds, fmt.Sprintf("%s.id = ?", table))
			args = append(args, filter.ID)
		}
		if filter.ShortID != "" {
			conds = append(conds, fmt.Sprintf("%s.short_id = ?", table))
			args = append(args, filter.ShortID)
		}
		if filter.RegularOrShortID != "" {
			conds = append(conds, fmt.Sprintf("%s.id = ? OR %s.short_id = ?", table, table))
			args = append(args, filter.RegularOrShortID, filter.RegularOrShortID)
		}
		if filter.UserID != "" {
			conds = append(conds, fmt.Sprintf("(%s.user_id = ? OR %s.user_id = '')", table, table))
			args = append(args, filter.UserID)
		}
		if filter.UserSource != "" {
			conds = append(conds, "users.source = ?")
			args = append(args, filter.UserSource)
		}
		if filter.Visibility != "" {
			conds = append(conds, fmt.Sprintf("%s.visibility = ?", table))
			args = append(args, filter.Visibility)
		}
		if filter.ApiToken != "" {
			conds = append(conds, fmt.Sprintf("%s.api_token = ?", table))
			args = append(args, filter.ApiToken)
		}
		if filter.FontEnabled != nil && *filter.FontEnabled == false {
			conds = append(conds, "enabled_fonts.id IS NULL")
		}

		if len(conds) != 0 {
			query += "WHERE " + strings.Join(conds, " AND ") + " "
		}

		if filter.Limit != 0 {
			query += "LIMIT ? "
			args = append(args, filter.Limit)
		}
		if filter.Offset != 0 {
			query += "OFFSET ? "
			args = append(args, filter.Offset)
		}
	}

	return query, args
}

type CountRow struct {
	Count int `db:"count"`
}
