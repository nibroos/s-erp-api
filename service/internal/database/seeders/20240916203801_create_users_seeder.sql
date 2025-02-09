BEGIN;

-- marketing, sales, purchasing, inventory, production, exim, accounting, beacukai
INSERT INTO
  users (
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
    'admin',
    'admin@yubipro.com',
    'Admin',
    crypt('adminyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'manager',
    'manager@yubipro.com',
    'Manager',
    crypt('manageryubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'marketing',
    'marketing@yubipro.com',
    'Marketing',
    crypt('marketingyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'sales',
    'sales@yubipro.com',
    'Sales',
    crypt('salesyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'purchasing',
    'purchasing@yubipro.com',
    'Purchasing',
    crypt('purchasingyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'inventory',
    'inventory@yubipro.com',
    'Inventory',
    crypt('inventoryyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'production',
    'production@yubipro.com',
    'Production',
    crypt('productionyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'exim',
    'exim@yubipro.com',
    'Exim',
    crypt('eximyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'accounting',
    'accounting@yubipro.com',
    'Accounting',
    crypt('accountingyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    'beacukai',
    'beacukai@yubipro.com',
    'Beacukai',
    crypt('beacukaiyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;