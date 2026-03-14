-- 011_business_details.sql
-- Configure business details in app_settings table
INSERT INTO app_settings (key, value) VALUES ('business_name', 'PARFUM TRADERS') ON CONFLICT (key) DO NOTHING;
INSERT INTO app_settings (key, value) VALUES ('business_gstin', '33AUSPR1909H1ZC') ON CONFLICT (key) DO NOTHING;
INSERT INTO app_settings (key, value) VALUES ('business_address_line1', 'No. 9/21, 1st floor, Sadiq Basha Nagar,') ON CONFLICT (key) DO NOTHING;
INSERT INTO app_settings (key, value) VALUES ('business_address_line2', '2nd Street, Virugambakkam, Chennai - 600092') ON CONFLICT (key) DO NOTHING;
INSERT INTO app_settings (key, value) VALUES ('business_phone', '7904769823') ON CONFLICT (key) DO NOTHING; 
