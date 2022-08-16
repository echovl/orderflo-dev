BEGIN;

DROP TABLE
  IF EXISTS users;

DROP TABLE
  IF EXISTS companies;

DROP TABLE
  IF EXISTS customers;

DROP TABLE
  IF EXISTS uploads;

DROP TABLE
  IF EXISTS public_fonts;

DROP TABLE
  IF EXISTS private_fonts;

DROP TABLE
  IF EXISTS enabled_fonts;

DROP TABLE
  IF EXISTS fonts;

DROP TABLE
  IF EXISTS frames;

DROP TABLE
  IF EXISTS projects;

DROP TABLE
  IF EXISTS components;

DROP TABLE
  IF EXISTS templates;

DROP TABLE
  IF EXISTS tags;

DROP TABLE
  IF EXISTS template_tags;

DROP TABLE
  IF EXISTS colors;

DROP TABLE
  IF EXISTS template_colors;

DROP TABLE
  IF EXISTS template_metadata;

DROP TABLE
  IF EXISTS subscription_plans;

DROP TABLE
  IF EXISTS subscription_plan_billings;

COMMIT;
