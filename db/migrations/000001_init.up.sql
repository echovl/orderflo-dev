BEGIN;

CREATE TABLE
  IF NOT EXISTS users (
    id VARCHAR(50),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    avatar VARCHAR(255) NOT NULL,
    company VARCHAR(255) NOT NULL,
    email_verified BOOL NOT NULL,
    phone_verified BOOL NOT NULL,
    role VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_token VARCHAR(1000) NOT NULL,
    source VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS uploads (
    id VARCHAR(50),
    name VARCHAR(255) NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    folder VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS fonts (
    id VARCHAR(50),
    full_name VARCHAR(255) NOT NULL,
    family VARCHAR(255) NOT NULL,
    postscript_name VARCHAR(255) NOT NULL,
    preview VARCHAR(255) NOT NULL,
    style VARCHAR(255) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    category VARCHAR(255) NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS enabled_fonts (
    id VARCHAR(50),
    user_id VARCHAR(50),
    font_id VARCHAR(50) NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS frames (
    id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    visibility VARCHAR(255) NOT NULL,
    width FLOAT NOT NULL,
    height FLOAT NOT NULL,
    unit VARCHAR(10) NOT NULL,
    preview VARCHAR(1000) NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS projects (
    id VARCHAR(255),
    short_id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    preview VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS components (
    id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    preview VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS templates (
    id VARCHAR(255),
    short_id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    published BOOL NOT NULL,
    preview VARCHAR(255),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS template_tags (
    id VARCHAR(255) NOT NULL,
    template_id VARCHAR(255) NOT NULL,
    tag VARCHAR(255) NOT NULL,
    position INT NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS template_colors (
    id VARCHAR(255) NOT NULL,
    template_id VARCHAR(255) NOT NULL,
    color VARCHAR(255) NOT NULL,
    position INT NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS template_metadata (
    id VARCHAR(255) NOT NULL,
    license VARCHAR(255) NOT NULL,
    orientation VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS subscription_plans (
    id VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    external_product_id VARCHAR(255) NOT NULL,
    auto_bill_outstanding BOOL NOT NULL,
    setup_fee VARCHAR(255) NOT NULL,
    max_templates INT NOT NULL,
    PRIMARY KEY (id)
  );

CREATE TABLE
  IF NOT EXISTS subscription_plan_billings (
    id VARCHAR(255) NOT NULL,
    `interval` VARCHAR(255) NOT NULL,
    price VARCHAR(255) NOT NULL,
    subscription_plan_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
  );

COMMIT;
