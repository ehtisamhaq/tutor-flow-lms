-- Seed Users
-- Password for all users is: password123

INSERT INTO
    users (
        id,
        email,
        password_hash,
        first_name,
        last_name,
        role,
        status,
        created_at,
        updated_at
    )
VALUES (
        gen_random_uuid (),
        'student@tutorflow.com',
        crypt (
            'password123',
            gen_salt ('bf')
        ),
        'John',
        'Student',
        'student',
        'active',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid (),
        'tutor@tutorflow.com',
        crypt (
            'password123',
            gen_salt ('bf')
        ),
        'Jane',
        'Tutor',
        'tutor',
        'active',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid (),
        'manager@tutorflow.com',
        crypt (
            'password123',
            gen_salt ('bf')
        ),
        'Mike',
        'Manager',
        'manager',
        'active',
        NOW(),
        NOW()
    ),
    (
        gen_random_uuid (),
        'admin@tutorflow.com',
        crypt (
            'password123',
            gen_salt ('bf')
        ),
        'Alice',
        'Admin',
        'admin',
        'active',
        NOW(),
        NOW()
    ) ON CONFLICT (email) DO NOTHING;
