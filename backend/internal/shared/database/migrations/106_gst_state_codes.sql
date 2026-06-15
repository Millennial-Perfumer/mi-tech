CREATE TABLE gst_state_codes (
    code VARCHAR(2) PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    aliases TEXT[] NOT NULL
);

-- Index for fast array overlap matching
CREATE INDEX idx_gst_state_codes_aliases ON gst_state_codes USING gin(aliases);

-- Seed all 37 States / UTs with standard variations
INSERT INTO gst_state_codes (code, name, aliases) VALUES
('01', 'Jammu & Kashmir', ARRAY['jammu & kashmir', 'jammu and kashmir', 'j&k', 'jk']),
('02', 'Himachal Pradesh', ARRAY['himachal pradesh', 'hp']),
('03', 'Punjab', ARRAY['punjab', 'pb']),
('04', 'Chandigarh', ARRAY['chandigarh', 'ch']),
('05', 'Uttarakhand', ARRAY['uttarakhand', 'uk', 'uttaranchal']),
('06', 'Haryana', ARRAY['haryana', 'hr']),
('07', 'Delhi', ARRAY['delhi', 'dl', 'ncr']),
('08', 'Rajasthan', ARRAY['rajasthan', 'rj']),
('09', 'Uttar Pradesh', ARRAY['uttar pradesh', 'up']),
('10', 'Bihar', ARRAY['bihar', 'br']),
('11', 'Sikkim', ARRAY['sikkim', 'sk']),
('12', 'Arunachal Pradesh', ARRAY['arunachal pradesh', 'ar']),
('13', 'Nagaland', ARRAY['nagaland', 'nl']),
('14', 'Manipur', ARRAY['manipur', 'mn']),
('15', 'Mizoram', ARRAY['mizoram', 'mz']),
('16', 'Tripura', ARRAY['tripura', 'tr']),
('17', 'Meghalaya', ARRAY['meghalaya', 'ml']),
('18', 'Assam', ARRAY['assam', 'as']),
('19', 'West Bengal', ARRAY['west bengal', 'wb']),
('20', 'Jharkhand', ARRAY['jharkhand', 'jh']),
('21', 'Odisha', ARRAY['odisha', 'orissa', 'or', 'od']),
('22', 'Chhattisgarh', ARRAY['chhattisgarh', 'cg', 'chhatisgarh']),
('23', 'Madhya Pradesh', ARRAY['madhya pradesh', 'mp']),
('24', 'Gujarat', ARRAY['gujarat', 'gj']),
('26', 'Dadra & Nagar Haveli and Daman & Diu', ARRAY['dadra and nagar haveli', 'daman and diu', 'dnhdd']),
('27', 'Maharashtra', ARRAY['maharashtra', 'mh']),
('29', 'Karnataka', ARRAY['karnataka', 'ka']),
('30', 'Goa', ARRAY['goa', 'ga']),
('31', 'Lakshadweep', ARRAY['lakshadweep', 'ld']),
('32', 'Kerala', ARRAY['kerala', 'kl']),
('33', 'Tamil Nadu', ARRAY['tamil nadu', 'tn', 'tamilnadu']),
('34', 'Puducherry', ARRAY['puducherry', 'pondicherry', 'py']),
('35', 'Andaman & Nicobar Islands', ARRAY['andaman & nicobar', 'andaman and nicobar', 'an']),
('36', 'Telangana', ARRAY['telangana', 'tg', 'ts']),
('37', 'Andhra Pradesh', ARRAY['andhra pradesh', 'ap']),
('38', 'Ladakh', ARRAY['ladakh', 'la']),
('97', 'Other Territory', ARRAY['other territory', 'ot']);
