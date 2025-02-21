ROLLBACK;

BEGIN;

-- marketing, sales, purchasing, inventory, production, exim, accounting, beacukai
INSERT INTO
  users (
    branch_id,
    username,
    email,
    name,
    password,
    address,
    created_at,
    updated_at
  )
VALUES
  (
    NULL,
    'admin',
    'admin@yubipro.com',
    'Admin',
    crypt('adminyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'manager',
    'manager@yubipro.com',
    'Manager',
    crypt('manageryubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'marketing',
    'marketing@yubipro.com',
    'Marketing',
    crypt('marketingyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'sales',
    'sales@yubipro.com',
    'Sales',
    crypt('salesyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'purchasing',
    'purchasing@yubipro.com',
    'Purchasing',
    crypt('purchasingyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'inventory',
    'inventory@yubipro.com',
    'Inventory',
    crypt('inventoryyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'production',
    'production@yubipro.com',
    'Production',
    crypt('productionyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'exim',
    'exim@yubipro.com',
    'Exim',
    crypt('eximyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'accounting',
    'accounting@yubipro.com',
    'Accounting',
    crypt('accountingyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    1,
    'beacukai',
    'beacukai@yubipro.com',
    'Beacukai',
    crypt('beacukaiyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;