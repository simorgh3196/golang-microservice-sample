-- テスト用テナントの挿入
INSERT INTO tenants (id, name, plan)
VALUES ('018f2b34-8c7a-7b3f-8000-000000000001', 'Acme Enterprise Corp', 'enterprise')
ON CONFLICT (id) DO NOTHING;

-- 'agf_live_test_key_123' の SHA-256 ハッシュ値で登録
-- echo -n "agf_live_test_key_123" | sha256sum -> 6f05676061e6be1e109cbef252c24d6e9603c346ab1457c108c1769c2453644d
INSERT INTO api_keys (id, tenant_id, key_hash, name, role, is_active)
VALUES (
    'key_test_01',
    '018f2b34-8c7a-7b3f-8000-000000000001',
    '6f05676061e6be1e109cbef252c24d6e9603c346ab1457c108c1769c2453644d',
    'Development Test Key',
    'admin',
    true
)
ON CONFLICT (id) DO NOTHING;
