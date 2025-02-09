BEGIN;

INSERT INTO
  company_profiles (
    company_name,
    company_address,
    company_phone,
    company_email,
    company_website,
    company_logo,
    company_description,
    company_remark,
    company_status,
    company_options_json,
    created_by_id,
    updated_by_id,
    deleted_by_id,
    created_at,
    updated_at
  )
VALUES
  (
    'Company',
    'Company Address',
    'Company Phone',
    'Company Email',
    'Company Website',
    'Company Logo',
    'Company Description',
    'Company Remark',
    1,
    '{}',
    1,
    1,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;