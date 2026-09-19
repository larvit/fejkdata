# Fake-data libraries: data-category inventory

Researched 2026-09-16. Sources: official docs and current `master`/`main`/`next` source trees. "Big four" = @faker-js/faker (JS), Faker (Python), gofakeit (Go), Datafaker (Java). Abbreviations in tables: **JS**, **PY**, **GO**, **DF**, **BG** (Bogus), **MM** (Mimesis), **CH** (Chance), **RB** (Ruby faker).

## 1. Category modules and their generators

Legend: names are the library's own identifiers (camelCase = JS/DF/BG, snake_case = PY/MM/RB, PascalCase = GO). Columns list what each library offers under that module; a `—` means the module does not exist there.

### person

| Lib | Generators |
|---|---|
| JS | firstName, lastName, middleName, fullName, prefix, suffix, sex, sexType, gender, bio, jobArea, jobDescriptor, jobTitle, jobType, zodiacSign (all name methods take `sex: female|male|generic`) |
| PY | first_name/_female/_male/_nonbinary, last_name/_female/_male/_nonbinary, name/_female/_male/_nonbinary, prefix/_female/_male/_nonbinary, suffix/_female/_male/_nonbinary, language_name; separate providers: job (job, job_female, job_male), ssn (ssn; per-locale variants: vat_id, itin, ein, nif, cpf, personnummer …), passport (passport_number, passport_dob, passport_owner), profile (profile → job, company, ssn, residence, current_location, blood_group, website, username, name, sex, address, mail, birthdate; simple_profile) |
| GO | Person (struct), Name, NamePrefix, NameSuffix, FirstName, MiddleName, LastName, Gender, Age, Ethnicity, SSN, EIN, Hobby, SocialMedia, Bio, Contact, Email, Phone, PhoneFormatted, Teams |
| DF | Name: name, nameWithMiddle, fullName, firstName, femaleFirstName, maleFirstName, lastName, prefix, suffix, title, username; Demographic: race, educationalAttainment, demonym, sex, maritalStatus; Gender; Pronouns; BloodType; Mbti; Zodiac; Relationship; Hobby; Mood; Job; Passport; DrivingLicense; IdNumber (valid/invalid per country: US SSN, SE, ZA, SG FIN/UIN, CN, PT NIF, MX, PL PESEL, KR RRN, GE); CPF/CNPJ (BR); Nigeria; Australia |
| BG | Name: FirstName, LastName, FullName, Prefix, Suffix, FindName, JobTitle, JobDescriptor, JobArea, JobType; Person object (Gender, FirstName, LastName, FullName, UserName, Avatar, Email, DateOfBirth, Address, Phone, Website, Company); country extensions: Ssn/Ein (US), Personnummer/Samordningsnummer (SE), Cpr (DK), Henkilotunnus (FI), Fodselsnummer (NO), Pesel/Nip/Regon (PL), CodiceFiscale (IT), Cpf/Cnpj (BR), Sin (CA), Nino (GB), Cnp (RO), NationalNumber (BE), Nif/Nipc (PT) |
| MM | Person: first_name, surname/last_name, patronymic, full_name, name, title, username, password, email, birthdate, gender, gender_symbol, gender_code, sex, height, weight, blood_type, occupation, nationality, university, academic_degree, language, phone_number, telephone, identifier |
| CH | first, last, name, prefix, suffix, gender, age, birthday, ssn, cpf, cf (IT codice fiscale), profession |
| RB | Name (name, name_with_middle, first_name, male_first_name, female_first_name, neutral_first_name, last_name, prefix, suffix, initials), Gender, Demographic, IDNumber (per-country), Job, Relationship, Blood, DrivingLicence, Avatar, ChileRut, SouthAfrica, NationalHealthService |

### address / location

| Lib | Generators |
|---|---|
| JS (location) | buildingNumber, cardinalDirection, city, continent, country, countryCode, county, direction, language, latitude, longitude, nearbyGPSCoordinate, ordinalDirection, postalAddress, secondaryAddress, state({abbreviated}), street, streetAddress, timeZone, zipCode({state, format}) |
| PY (address) | address, building_number, city, city_suffix, country, country_code, current_country, current_country_code, postcode, street_address, street_name, street_suffix; locale extras e.g. en_US: city_prefix, secondary_address, administrative_unit, state_abbr, zipcode_plus4, postcode_in_state, military_ship/state/apo/dpo; ja_JP: prefecture, city, town, chome, ban, gou, building_name; fr_FR: department, department_name, department_number, region; es_ES: region/autonomous_community; de_DE: city_with_postcode, street_suffix_short/long; en_GB: county; separate **geo** provider: coordinate, latitude, longitude, latlng, local_latlng(country), location_on_land (real geonames.org places with tz) |
| GO (address) | Address (struct), City, Country, CountryAbr, State, StateAbr, Street, StreetName, StreetNumber, StreetPrefix, StreetSuffix, Unit, Zip, Latitude, LatitudeInRange, Longitude, LongitudeInRange |
| DF (Address) | streetName, streetAddressNumber, streetAddress, secondaryAddress, zipCode, postcode, eircode, zipCodePlus4, zipCodeByState, countyByZipCode, streetSuffix, streetPrefix, citySuffix, cityPrefix, city, cityName, state, stateAbbr, latitude, longitude, latLon, lonLat, timeZone, country, countryCode, buildingNumber, fullAddress, mailBox; plus Country (name, code2/3, capital, currency, flag), Nation (nationality, language, capital, flag), Compass, Mountain, Planet, Space, Locality (locale codes) |
| BG (Address) | ZipCode, City, StreetAddress, CityPrefix, CitySuffix, StreetName, BuildingNumber, StreetSuffix, SecondaryAddress, County, Country, FullAddress, CountryCode, State, StateAbbr, Latitude, Longitude, Direction, CardinalDirection, OrdinalDirection; premium Bogus.Locations: GPS, altitude, depth, geohash |
| MM (Address) | street_number, street_name, street_suffix, secondary_address, address, state/region/province/federal_subject/prefecture (abbr), postal_code/zip_code, country, country_code (A2/A3/numeric), country_emoji_flag, default_country, city, latitude, longitude, coordinates (DMS option), continent, calling_code/isd_code, iata_code, icao_code |
| CH | address, street, city, zip, postal (CA), postcode (GB), state, province, country, areacode, phone, latitude, longitude, altitude, depth, coordinates, geohash, locale |
| RB | city, street_name, street_address, secondary_address, building_number, mail_box, community, zip_code(state_abbreviation:), zip, postcode, time_zone, street_suffix, city_suffix, city_prefix, state, state_abbr, country, country_by_code, country_name_to_code, country_code, country_code_long, latitude, longitude, full_address, full_address_as_hash; Travel: airport, train_station; Locations: australia |

### company

| Lib | Generators |
|---|---|
| JS | name, buzzAdjective, buzzNoun, buzzPhrase, buzzVerb, catchPhrase, catchPhraseAdjective, catchPhraseDescriptor, catchPhraseNoun (locale data: adjective, descriptor, noun, legal_entity_type, name_pattern) |
| PY | company, company_suffix, catch_phrase, bs; locale extras (e.g. it_IT/pl_PL company_vat, nl_NL, ru_RU large_company, etc.) |
| GO | Company, CompanySuffix, BS, Blurb, BuzzWord, Slogan, Job, JobDescriptor, JobLevel, JobTitle |
| DF | Company: name, suffix, industry, profession, buzzword, catchPhrase, bs, logo, domainName, url; Business; Brand; IndustrySegments; Marketing; Team; Restaurant; Subscription; Twitter; University; Educator |
| BG | CompanySuffix, CompanyName, CatchPhrase, Bs |
| MM | Finance.company, company_type |
| CH | company, profession |
| RB | Company (name, suffix, industry, profession, type, catch_phrase, buzzword, bs, logo, ein, duns_number, swedish_organisation_number, czech_organisation_number, french_siren/siret, norwegian/australian/spanish/polish/russian/south_african/brazilian ids …), Business, Marketing, IndustrySegments, Team, University, Educator, Restaurant, Subscription, Construction |

### internet

| Lib | Generators |
|---|---|
| JS | displayName, domainName, domainSuffix, domainWord, email, exampleEmail, emoji, httpMethod, httpStatusCode, ip, ipv4, ipv6, jwt, jwtAlgorithm, mac, password, port, protocol, url, userAgent, username |
| PY | email, safe_email, free_email, company_email, ascii_* variants, free_email_domain, domain_name, domain_word, safe_domain_name, tld, hostname, dga, http_method, http_status_code, iana_id, image_url, ipv4, ipv4_network_class, ipv4_private, ipv4_public, ipv6, mac_address, nic_handle(s), port_number, ripe_id, slug, uri, uri_extension, uri_page, uri_path, url, user_name; user_agent provider (chrome, firefox, safari, opera, internet_explorer, platform tokens); emoji provider |
| GO | URL, UrlSlug, DomainName, DomainSuffix, IPv4Address, IPv6Address, MacAddress, HTTPStatusCode(Simple), HTTPMethod, HTTPVersion, LogLevel, UserAgent (+Chrome/Firefox/Opera/Safari/API), Username, Password, Emoji (+22 category fns), InputName, Svg |
| DF | Internet: username, emailAddress(name), safeEmailAddress, emailSubject, domainName/Word/Suffix, url, webdomain, image, httpMethod, password(opts), port, macAddress, ipV4Address, privateIpV4Address, publicIpV4Address, ipV4Cidr, ipV6Address, ipV6Cidr, slug, uuidv3/4/7, userAgent, botUserAgent; Http; Domain; Sip; Emoji; SlackEmoji; Credentials; Aws; Azure; Hashing; Fingerprint; Computer; Device; Camera; Drone; Robin |
| BG | Avatar, Email, ExampleEmail, UserName, UserNameUnicode, DomainName, DomainWord, DomainSuffix, Ip, Port, IpAddress, IpEndPoint, Ipv6, UserAgent, Mac, Password, Color, Protocol, Url, UrlWithPath, UrlRootedPath |
| MM | Internet: content_type, dsn, http_status_message/code, http_method, ip_v4/v6 (+object, cidr, with_port, special), cloud_region, asn, mac_address, hostname, url, uri, query_string/parameters, tld, user_agent, port, path, slug, public_dns, http_request/response_headers |
| CH | avatar, color, domain, email, fbid, google_analytics, hashtag, ip, ipv6, klout, tld, twitter, url |
| RB | Internet (email, username, password, domain_name, ip_v4/v6, mac_address, url, slug, user_agent, uuid, bot_user_agent …), Internet::HTTP, Omniauth, Stripe, X (Twitter), Avatar, Placeholdit, LoremFlickr |

### finance / payment / bank

| Lib | Generators |
|---|---|
| JS | accountName, accountNumber, amount, bic, bitcoinAddress, creditCardCVV, creditCardIssuer, creditCardNumber, currency, currencyCode, currencyName, currencyNumericCode, currencySymbol, ethereumAddress, iban, litecoinAddress, pin, routingNumber, transactionDescription, transactionType |
| PY | bank: aba, bank, bank_country, bban, iban, swift, swift8, swift11; credit_card: credit_card_number, credit_card_provider, credit_card_expire, credit_card_security_code, credit_card_full; currency: currency, currency_code, currency_name, currency_symbol, cryptocurrency(_code/_name), pricetag |
| GO | Price, CreditCard, CreditCardCvv, CreditCardExp, CreditCardNumber, CreditCardType, Currency, CurrencyLong, CurrencyShort, AchRouting, AchAccount, BitcoinAddress, BitcoinPrivateKey, BankName, BankType, Cusip, Isin |
| DF | Finance: creditCard(type), bic, iban(country), ibanSupportedCountries, usRoutingNumber; Money: currency, currencyCode, currencyNumericCode, currencySymbol; Currency; Stock; FinancialTerms; CryptoCoin; Coin; Business |
| BG | Account, AccountName, Amount, TransactionType, Currency, CreditCardNumber, CreditCardCvv, BitcoinAddress, EthereumAddress, RoutingNumber, Bic, Iban; GB SortCode, VatNumber |
| MM | Finance: bank, currency_iso_code, currency_symbol, cryptocurrency_iso_code/_symbol, price, price_in_btc, stock_ticker, stock_name, stock_exchange; Payment: cid, bitcoin_address, ethereum_address, credit_card_network/number/expiration_date, cvv, credit_card_owner |
| CH | cc, cc_type, currency, currency_pair, dollar, euro, exp, exp_month, exp_year |
| RB | Finance (credit_card, vat_number, ticker, stock_market), Bank (name, swift_bic, iban, account_number, routing_number, bsb_number), Currency, Crypto, Blockchain (bitcoin, ethereum, tezos, aeternity), Invoice, Stripe, Coin |

### commerce / product

| Lib | Generators |
|---|---|
| JS | department, isbn, price, product, productAdjective, productDescription, productMaterial, productName, upc |
| PY | barcode: ean, ean8, ean13, localized_ean*; isbn: isbn10, isbn13; sbn; (no product/commerce provider in core; community Ecommerce provider) |
| GO | Product (struct), ProductName, ProductDescription, ProductCategory, ProductFeature, ProductMaterial, ProductUPC, ProductAudience, ProductDimension, ProductUseCase, ProductBenefit, ProductSuffix, ProductISBN |
| DF | Commerce: department, productName, material, brand, vendor, price(min,max), promotionCode; Barcode; Code (isbn10/13, gtin8/13, ean, asin, imei); Appliance; GarmentSize; Size; Tire; House; ElectricalComponents |
| BG | Department, Price, Categories, ProductName, Color, Product, ProductAdjective, ProductMaterial, Ean8, Ean13 |
| MM | Code: locale_code, issn, isbn, ean, imei, pin |
| CH | — |
| RB | Commerce (color, department, material, product_name, price, promotion_code, brand, vendor), Barcode, Code (isbn, ean, asin, imei, npi, nric, rut, sin), Appliance, Device, Camera, House |

### vehicle / transport

| Lib | Generators |
|---|---|
| JS | bicycle, color, fuel, manufacturer, model, type, vehicle, vin, vrm; airline: aircraftType, airline, airplane, airport, flightNumber, recordLocator, seat |
| PY | automotive: license_plate, vin (per-locale plate formats) |
| GO | Car (struct), CarMaker, CarModel, CarType, CarFuelType, CarTransmissionType; Airline*: AircraftType, Airplane, Airport, AirportIATA, FlightNumber, RecordLocator, Seat |
| DF | Vehicle: vin, manufacturer, make, model(make), makeAndModel, style, color, upholstery(+Color/Fabric), transmission, driveType, fuelType, carType, engine, carOptions, standardSpecs, doors, licensePlate(state); Aviation; Transport; DrivingLicense |
| BG | Vin, Manufacturer, Model, Type, Fuel; GbRegistrationPlate |
| MM | Transport: manufacturer, car, airplane, vehicle_registration_code(locale) |
| CH | — |
| RB | Vehicle (vin, manufacture, make, model, make_and_model, style, color, transmission, drive_type, fuel_type, car_type, engine, car_options, standard_specs, doors, door, year, mileage, license_plate, singapore_license_plate, version), Drone, Travel::Airport, Travel::TrainStation |

### lorem / text / words

| Lib | Generators |
|---|---|
| JS | lorem: lines, paragraph(s), sentence(s), slug, text, word(s); word: adjective, adverb, conjunction, interjection, noun, preposition, sample, verb, words; hacker: abbreviation, adjective, ingverb, noun, phrase, verb; string: alpha, alphanumeric, binary, fromCharacters, hexadecimal, nanoid, numeric, octal, sample, symbol, ulid, uuid |
| PY | lorem: word(s), sentence(s), paragraph(s), text(s), get_words_list; misc: password, md5, sha1, sha256, binary, boolean, null_boolean, csv/dsv/psv/tsv, fixed_width, json, json_bytes, image, tar, zip, uuid4 |
| GO | Word families: Noun* (Common/Concrete/Abstract/Collective*/Countable/Uncountable/Proper/Determiner), Verb* (Action/Linking/Helping/Transitive/Intransitive), Adverb* (Manner/Degree/Place/Time*/Frequency*), Preposition* (Simple/Double/Compound), Adjective* (8 kinds), Pronoun* (8 kinds), Connective* (6 kinds), Word, Interjection, Sentence, Paragraph, LoremIpsum{Word,Sentence,Paragraph}, Question, Quote, Phrase{,Noun,Verb,Adverb,Preposition}, Comment; Hacker*, Hipster{Word,Sentence,Paragraph}; Letter(N), Digit(N), Vowel, Lexify, Numerify, RandomString, ShuffleStrings |
| DF | Lorem (word(s), sentence(s), paragraph(s), characters, fixedString, maxLengthSentence, supplemental), Text (character, upper/lowercase, text with symbol rules), Verb, Word, Hacker, Hipster, Shakespeare, Yoda, Joke, ChuckNorris, Matz, NatoPhoneticAlphabet, FunnyName, Marketing |
| BG | Lorem: Word(s), Letter, Sentence(s), Paragraph(s), Text, Lines, Slug; Hacker: Abbreviation, Adjective, Noun, Verb, IngVerb, Phrase; Rant: Review(s) |
| MM | Text: alphabet, level, text, sentence, title, words, word, quote, color, hex_color, rgb_color, answer, emoji |
| CH | paragraph, sentence, syllable, word, character, letter, string |
| RB | Lorem (word(s), character(s), sentence(s), paragraph(s), question(s), paragraph_by_chars, multibyte), Markdown, Hipster, Hacker, Adjective, Verbs, Quote, Quotes::Shakespeare/Chiquito/Rajnikanth, ChuckNorris, Emotion, Source (code snippets), Html, Json, Types, Alphanumeric, String, Boolean, NatoPhoneticAlphabet |

### date / time

| Lib | Generators |
|---|---|
| JS | anytime, between, betweens, birthdate({mode: age|year}), future, past, recent, soon, month, weekday, timeZone (locale data: month, weekday) |
| PY | date_time: date, date_time, date_object, time, time_object, date_between(_dates), date_time_between(_dates), date_this_century/decade/month/year, date_time_this_*, date_time_ad, date_of_birth(minimum_age, maximum_age), future_date(time), past_date(time), iso8601, unix_time, time_delta, time_series, timezone, pytimezone, am_pm, century, year, month, month_name, day_of_month, day_of_week |
| GO | Date, PastDate, FutureDate, DateRange, NanoSecond, Second, Minute, Hour, Month, MonthString, Day, WeekDay, Year, TimeZone, TimeZoneAbv, TimeZoneFull, TimeZoneOffset, TimeZoneRegion |
| DF | DateAndTime: future, past, between, birthday, birthdayLocalDate, duration, period; Time; TimeAndDate |
| BG | Past, PastOffset, Soon, SoonOffset, Future, FutureOffset, Between, BetweenOffset, Recent, RecentOffset, Timespan, Month, Weekday |
| MM | Datetime: date, datetime, time, timestamp, formatted_*, week_date, day_of_week, month, year, day_of_month, timezone, gmt_offset, periodicity, future/past_date(time), duration, bulk_create_datetimes |
| CH | ampm, date, hammertime, hour, millisecond, minute, month, second, timestamp, timezone, weekday, year |
| RB | Date (between, between_except, forward, backward, birthday, in_date_period, on_day_of_week_between), Time (between, between_dates, forward, backward) |

### phone

| Lib | Generators |
|---|---|
| JS | number({style: human|national|international}), imei |
| PY | phone_number, country_calling_code, msisdn (per-locale formats) |
| GO | Phone, PhoneFormatted |
| DF | PhoneNumber: cellPhone, cellPhoneInternational, phoneNumber, phoneNumberInternational, phoneNumberNational, extension, subscriberNumber |
| BG | PhoneNumber, PhoneNumberFormat |
| MM | Person.phone_number/telephone (mask), Address.calling_code |
| CH | phone, areacode |
| RB | PhoneNumber (phone_number, cell_phone, country_code, phone_number_with_country_code, cell_phone_with_country_code, cell_phone_in_e164, area_code, exchange_code, subscriber_number, extension) |

### science / medical

| Lib | Generators |
|---|---|
| JS | chemicalElement, unit; medical locale data (en only) |
| PY | — (community: Healthcare, Biology, Geoscience, Scientific) |
| GO | — |
| DF | Science: element, elementSymbol, unit, scientist, tool, quark, leptons, bosons; Medical, Medication, Disease, MedicalProcedure, CareProvider, Observation, BloodType, Measurement, Weather, Planet, Space, Mountain, Cannabis, LargeLanguageModel |
| BG | premium Bogus.Healthcare |
| MM | Science: rna_sequence, dna_sequence |
| CH | — |
| RB | Science (element, element_symbol, element_state, element_subcategory, scientist, modifier, tool), Space, Measurement, Cannabis, Medical (NationalHealthService) |

### music / animal / food / books / entertainment

| Lib | Generators |
|---|---|
| JS | music: album, artist, genre, songName; animal: bear, bird, cat, cetacean, cow, crocodilia, dog, fish, horse, insect, lion, petName, rabbit, rodent, snake, type; food: adjective, description, dish, ethnicCategory, fruit, ingredient, meat, spice, vegetable; book: author, format, genre, publisher, series, title |
| PY | — (core has none; community Music, Sci Fi) |
| GO | Song, SongName, SongArtist, SongGenre; PetName, Animal, AnimalType, FarmAnimal, Cat, Dog, Bird; Fruit, Vegetable, Breakfast, Lunch, Dinner, Snack, Dessert, Drink; Beer{Alcohol,Blg,Hop,Ibu,Malt,Name,Style,Yeast}; Book, BookTitle, BookAuthor, BookGenre; Movie, MovieName, MovieGenre; Celebrity{Actor,Business,Sport}; Minecraft (18 fns) |
| DF | Music: instrument, key, chord, genre; RockBand; Artist; Kpop; Hololive; Animal: name, scientificName, genus, species; Cat, Dog, Horse; Food: ingredient, allergen, spice, dish, fruit, vegetable, sushi, measurement; Apple, Beer, Cheese, Coffee, Dessert, IceCream, Tea; Book, Movie, Show, OscarMovie; ~75 entertainment franchises, ~31 videogames, 9 sports |
| BG | Music (premium Hollywood: movies, TV, actors) |
| MM | Food: vegetable, fruit, dish, spices, drink |
| CH | animal (with type), tv, radio, rpg, dice |
| RB | Music (+11 bands), Creature (animal, bird, cat, dog, horse), Food, Beer, Coffee, Tea, Dessert, Book, Books (4 series), Movie/Movies (17), TvShows (39), Games (26), JapaneseMedia (10), Sports (6), Fantasy::Tolkien, Religion::Bible, Superhero, DcComics, Kpop, Ancient, GreekPhilosophers, Cosmere |

### color / image / system / database / misc

| Lib | Generators |
|---|---|
| JS | color: cmyk, colorByCSSColorSpace, cssSupportedFunction, cssSupportedSpace, hsl, human, hwb, lab, lch, rgb, space; image: avatar, avatarGitHub, dataUri, personPortrait, url, urlLoremFlickr, urlPicsumPhotos; system: commonFileExt/Name/Type, cron, directoryPath, fileExt/Name/Path/Type, mimeType, networkInterface, semver; database: collation, column, engine, mongodbObjectId, type; git: branch, commitDate, commitEntry, commitMessage, commitSha; number: bigInt, binary, float, hex, int, octal, romanNumeral; datatype.boolean; science.unit |
| PY | color: color, color_name, safe_color_name, hex_color, safe_hex_color, rgb_color, rgb_css_color, color_hsl/hsv/rgb/rgb_float; file: file_extension, file_name, file_path, mime_type, unix_device, unix_partition; python: pybool, pydecimal, pydict, pyfloat, pyint, pyiterable, pylist, pyobject, pyset, pystr, pystr_format, pystruct, pytuple, enum; doi; emoji; user_agent |
| GO | Color, HexColor, RGBColor, HSLColor, SafeColor, NiceColors; Image, ImageJpeg, ImagePng; CSV, JSON, XML, SQL, FileExtension, FileMimeType, Template, Markdown, EmailText, FixedWidth; ID, UUID; Error* (9 kinds); AppName, AppVersion, AppAuthor; Language, LanguageAbbreviation, LanguageBCP, ProgrammingLanguage; School; Gamertag, Dice; Bool, Weighted, FlipACoin; Number/Int*/Uint*/Float* ranges |
| DF | Color, Image, File, App, ProgrammingLanguage, LanguageCode, Number, Bool, Unique, Options, Photography, Military, OlympicSport, Weather, Construction, Community, Chiquito … (263 providers total) |
| BG | Images: DataUri, PicsumUrl, PlaceholderUrl, LoremFlickrUrl; System: FileName, DirectoryPath, FilePath, CommonFileName, MimeType, CommonFileType/Ext, FileType, FileExt, Semver, Version, Exception, AndroidId, ApplePushToken, BlackBerryPin; Database: Column, Type, Collation, Engine; Randomizer (numbers, chars, Guid, Hash, Enum, WeightedRandom, Shuffle …) |
| MM | Hardware: resolution, screen_size, cpu, cpu_frequency, generation, cpu_codename, ram_type, ram_size, ssd_or_hdd, graphics, manufacturer, phone_model; Development: software_license, calver, version, stage, programming_language, os, boolean, system_quality_attribute; File, BinaryFile, Path, Cryptographic (uuid, hash, token, mnemonic), Numeric, Choice |
| CH | bool, falsy, floating, integer, natural, prime, hex, guid, hash, coin, dice, normal (Gaussian), n, unique, weighted, android_id, apple_token, bb_pin, wp7_anid, wp8_anid2 |
| RB | Color, File, Computer, Device, ProgrammingLanguage, App, Number, Boolean, Hash/Crypto, Json, Html, Markdown, Military, Compass, Coin, Mountain, Nation, Space, Time, Types, Slack Emoji, Vulnerability identifier |

Module presence summary:

| Module | JS | PY | GO | DF | BG | MM | CH | RB |
|---|---|---|---|---|---|---|---|---|
| person | x | x | x | x | x | x | x | x |
| address | x | x | x | x | x | x | x | x |
| company | x | x | x | x | x | (finance) | x | x |
| internet | x | x | x | x | x | x | x | x |
| finance | x | x | x | x | x | x | x | x |
| commerce/product | x | barcode only | x | x | x | code only | — | x |
| vehicle | x | plate/vin | x | x | x | x | — | x |
| airline | x | — | x | x | — | iata/icao | — | x |
| lorem/words | x | x | x | x | x | x | x | x |
| date | x | x | x | x | x | x | x | x |
| phone | x | x | x | x | x | x | x | x |
| science | x | — | — | x | — | dna | — | x |
| medical | data only | — | — | x | premium | — | — | x |
| music | x | — | x | x | — | — | — | x |
| animal | x | — | x | x | — | — | x | x |
| food | x | — | x | x | — | x | — | x |
| book | x | isbn | x | x | — | — | — | x |
| color | x | x | x | x | x | text | x | x |
| image | x | x | x | x | x | binaryfile | avatar | x |
| system/file | x | x | x | x | x | x | — | x |
| database | x | — | sql | — | x | dsn | — | — |
| git | x | — | — | — | — | — | — | — |
| hacker/hipster | x | — | x | x | x | — | — | x |
| national IDs | — | ssn per locale | SSN/EIN | IdNumber | ext. pkgs | identifier | ssn/cpf/cf | IDNumber |
| pop-culture | — | — | minecraft, movie | ~110 | premium | — | tv, rpg | ~85 |
| hardware | — | — | — | Computer/Device | — | x | — | Computer/Device |
| weather/space | — | — | — | x | — | — | — | Space |

## 2. Locale coverage and address data structure

### Locale counts

| Lib | Locales | Structure |
|---|---|---|
| JS | 77 dirs under `src/locales` (incl. `base`, `en_BORK`, `en_AU_ocker`) | One dir per locale, one file per module per key (`en/location/city_name.ts`). Fallback chain per Faker instance: `[de_CH, de, en, base]`; first locale that defines a key wins. `base` holds locale-independent data (ISO codes, time zones). A key may be explicitly `null` to mean "not applicable" (e.g. no zip codes for HK) so it does not fall through to `en`. https://fakerjs.dev/guide/localization.html |
| PY | 126 locale codes listed; per provider: address 67, person 86, ssn 56 | Python subclasses per `providers/<provider>/<locale>/__init__.py` overriding tuples/formats of the base. Missing provider for a locale falls back to `en_US`. Coverage is uneven (`am_ET` = phone only). Multi-locale `Faker(['it_IT','en_US'])` with optional weights. https://faker.readthedocs.io/en/master/locales.html |
| GO | 1 (English/US). No locale API; `Language`/`LanguageBCP` only emit language codes | Static Go maps in `data/*.go`. Open issue #352 "generate data for specific country/language" unanswered. https://github.com/brianvoe/gofakeit/issues/352 |
| DF | 127 yml files in `src/main/resources` (incl. 49 country-only files like `_SE.yml` and 78 language(-region) files); README says "60+" | One YAML per locale (`sv-SE.yml`) under `sv-SE: faker: address: …`; `en/` split into 264 per-provider files. Locale chain `de_CH → de → en`, key-by-key. Values are templates `#{Name.first_name}`. Custom yml via `faker.addPath(locale, path)`. Country file (`_SE.yml`) layered over language file. |
| BG | 50 (`data/*.locale.json`; README says 46) | Verbatim copy of faker.js (v5-era) locale JSON, merged with `data_extend/*.locale.json` by a gulp task, shipped as BSON. Falls back to `en`. https://github.com/bchavez/Bogus/wiki/Creating-Locales |
| MM | 47 (53 dirs incl. `global`, `int`, `bin`, template) | Per locale six JSON files: `address, datetime, finance, food, person, text`. Everything else (internet, payment, code, hardware…) is locale-independent. |
| CH | 1 (en; `postal` CA and `postcode` GB formats exist, `locale()` only returns codes) | Data inline in `chance.js`. |
| RB | 58 yml files ("over 40" in README) | One YAML per locale `lib/locales/<code>.yml` (`en` split into `en/*.yml`), I18n gem, key-level fallback to `en`. |

### Address data: real or synthetic, and correlation

| Lib | Cities | Street names | Postcodes | Hierarchy / correlation |
|---|---|---|---|---|
| JS | `en`: real list `city_name` (~1000 US cities) is one of 5 `city_pattern`s; the other 4 are `{prefix} {firstName}{suffix}` style synthetic. `de`: 327 real cities; `sv`: no `city_name`, pattern-only synthetic. Varies per locale. | `en`: `street_pattern` = `{firstName} {street_suffix}` / `{lastName} …` (synthetic); `de`: 1800 real street names (from one NRW region). | Format masks (`#####`, `#####-####`). `en_US` has `postcode_by_state` so `zipCode({state:'CA'})` gives a state-prefixed range. | None between city↔state↔zip. `county` is a flat list. `timeZone` is per-locale. `nearbyGPSCoordinate` is the only "correlated" geo function. Issue #983: en_CA returned US cities; PR #2141 added real cities for ZA locales — real-city coverage is locale-by-locale volunteer work. |
| PY | Per locale. `en_US`/`en_GB`: synthetic (`{city_prefix} {first_name}{city_suffix}`). `de_DE` 400+ real, `sv_SE` 45 real, `pl_PL` real, `ja_JP` real prefectures/cities/towns, `es_ES` real provinces + regions. | `en_US`: `{first_name} {street_suffix}`; `sv_SE`: prefix+suffix ("Björkgatan"); `de_DE`: synthetic name+Straße; `pl_PL`: real street-name pools. | Masks. `en_US`: `postcode_in_state` with real per-state numeric ranges (still random within range, "not correlated with cities"). `fr_FR`: postcode built from a real department number (correlated to department, not to city). `en_GB`: real outward-area letters but random assembly. `sv_SE`: `%####`. | No country→state→city hierarchy anywhere; `ja_JP` has prefecture–city pairs in data but methods pick independently. `geo.local_latlng(country)` returns real geonames places (name, lat, lon, country, tz) — the only real, correlated location record in the library. `address_formats` are weighted (`25.0` standard vs `1.0` military). |
| GO | Real list of major US cities | Real-looking suffix lists; StreetName is a pool | `#####` mask | None: `Address()` picks city, state, zip independently; lat/lon are uniform random on the globe, not near the city. Open issue #196 "Better address" unanswered. |
| DF | `en`: `city` formats are all synthetic (`#{city_prefix} #{Name.first_name}#{city_suffix}`); `cityName` draws from a separate real list. `sv-SE`: real `city_name` list; `de`: synthetic. | `en`: `#{Name.first_name} #{street_suffix}`; `de`: real `street_root` list combined randomly. | Masks; `en-US` `postcode_by_state` (`350##`), `zipCodeByState`, `countyByZipCode` (en-US only), `eircode` (IE). | None by default. Issue #1551 (Locale.CHINA yields "福建省南京市") open as "proposal". Issue #1477 unresolved directive for city names in some locales. |
| BG | Whatever faker.js v5 had per locale (`en`: synthetic patterns only — real `city_name` list postdates the copy) | synthetic | Masks; `en_US` `postcode_by_state` | None. Issue #481 "Invalid Zip Codes (US)" open. Issue #342 "City returns people's names" (side effect of `{firstName}` patterns). |
| MM | Real lists per locale (`en`: real US cities; `sv`: real municipalities incl. small towns) | `en`: real San Francisco street names; `sv`: real Swedish street names | `postal_code_fmt` mask per locale; `sv` has none | `state` (name+ISO 3166-2 abbr) real, but no link to city or postcode; `default_country` tied to locale. Open question #1584 on choosing default city per locale. |
| CH | Synthetic (3-syllable word) | Synthetic word + real suffix | Real-format masks | None |
| RB | `en`: synthetic patterns (`city_with_state` exists but is `city, state` random) ; some locales have real lists | `{first_name} {street_suffix}` | `#####`; `zip_code(state_abbreviation:)` uses per-state pattern (`900##`); doc says "may not be an actual state zip" | `full_address_as_hash` returns fields but does not correlate them. Issue #2881 (state vs state_abbr mismatch, country vs country_code mismatch) open; #1958 asked for geocodable city+zip; #2899 fake zips closed "not planned". |

**Which libraries correlate any address fields at all:** Python Faker (`fr_FR` postcode↔department, `en_US` zip↔state range, `geo.local_latlng` real records), faker-js/Datafaker/Bogus/Ruby (`zip↔state` range only, en_US), Mimesis (locale↔country only). None ships a country→region→city→postcode hierarchy or city↔lat/lon.

### Other localized fields (what varies per locale)

- JS: `sv` overrides color, commerce, company, date, internet, location, person, phone_number, team; `en` defines 24 modules; `base` 9. So finance/vehicle/animal/food/science are effectively English-only.
- PY: per-locale providers exist for address, automotive (plate formats), bank (IBAN/BBAN/SWIFT by country), company (VAT IDs), currency, date_time (month/day names), internet (TLDs, free email domains, user-name formats), job, lorem (word lists), person, phone_number, ssn (national IDs with check digits), passport, color, barcode (EAN prefixes), misc.
- DF `de.yml` overrides: address, company, compass, creature, lorem, hipster, name, color, commerce, book, university, chuck_norris, space, music, games, food, simpsons, dr_who, vehicle. Country files (`_SE.yml`) mainly phone_number and address formats.
- MM: only address, datetime, finance, food, person, text are localized.

### Name weighting

| Lib | Weighted by frequency? |
|---|---|
| PY | Yes, default `use_weighting=True`. `en_US`: ~500 female + ~500 male first names weighted from SSA decade tables, ~1000 surnames weighted by US Census; `sv_SE`: ~1000/1000/1000 weighted from Skatteverket/ISOF. Plus weighted `address_formats`, `random_element(OrderedDict)`. |
| JS | No. Plain lists (`en`: ~650 female, ~550 male, ~400 generic first names). `helpers.weightedArrayElement` exists for user data only. |
| GO | No. ~1500 first, ~500 last, plain slices. `Weighted(options, weights)` for user data. |
| DF | No. `en`: ~1200 male, ~4400 female first names, ~1000 last, unweighted. |
| BG | No (faker.js data); `Randomizer.WeightedRandom` for user data. |
| MM | No. `en`: ~4000 female, ~3000 male, ~2500 surnames, unweighted. |
| CH | No; `weighted()` helper only. |
| RB | No. |

## 3. Common complaints (issues and posts)

- **Uncorrelated address parts** — the single most repeated request across every library:
  - faker-ruby #1958 "Need an address object that has a street, city, zip that correlate" (geocoding breaks) https://github.com/faker-ruby/faker/issues/1958; #2881 "Consistent addresses" (state ≠ state_abbr, country ≠ country_code) https://github.com/faker-ruby/faker/issues/2881; #2899 "Fake Zip Codes" closed not-planned https://github.com/faker-ruby/faker/issues/2899; #275 zips like 90099 https://github.com/stympy/faker/issues/275
  - fzaninotto/Faker #625 "Faking Legitimate Addresses" https://github.com/fzaninotto/Faker/issues/625
  - java-faker #378 "Italy, Lima" / "USA, Moscow" https://github.com/DiUS/java-faker/issues/378; Datafaker #1551 wrong province–city for China https://github.com/datafaker-net/datafaker/issues/1551
  - gofakeit #196 "city may not be in the same state as the zip" https://github.com/brianvoe/gofakeit/issues/196
  - Bogus #481 invalid US zips https://github.com/bchavez/Bogus/issues/481
  - Python Faker #697 "city in country" https://github.com/joke2k/faker/issues/697; #1812 `states_postcode` fails for territories
- **Synthetic city names** ("East Jarretmouth", "Nord Müller-stadt"): faker-js #2021 merged `city`/`cityName` and documented that only some patterns are real https://github.com/faker-js/faker/issues/2021; Bogus #342 `Address.City` returns people's names https://github.com/bchavez/Bogus/issues/342; faker-js PR #2141 / #2127 / #3792 adding "real" or "more realistic" cities per locale.
- **Localized data falling through to English**: faker-js #983 en_CA cities were US cities https://github.com/faker-js/faker/issues/983; Datafaker #1477 unresolved directive for city names https://github.com/datafaker-net/datafaker/issues/1477; #1358 ru-MD city names; Python #2432 vi_VN outdated administrative units.
- **Fields within one record don't match**: Python #1185 profile name vs username/email https://github.com/joke2k/faker/issues/1185; #1420 email consistent with name https://github.com/joke2k/faker/issues/1420 (both stale, no maintainer answer); gofakeit `Person()` picks gender and first name independently and email from a fresh name.
- **No frequency weighting / flat distributions**: blood types come out uniform when O+ and A+ are >70% of population (https://www.statology.org/how-to-validate-enhance-faker-profile-data-generation/); Python's `use_weighting` is the only built-in answer and only some locales supply weights; #1815 weighting silently dropped after adding a provider https://github.com/joke2k/faker/issues/1815.
- **Formats that produce impossible values**: Python #1849 US phone numbers that can't exist; #1868 fr_FR postcodes with <5 digits; faker-js #1159 unrealistic BIC; #3429 routing numbers fixed to use a real Federal Reserve lookup table; Ruby #1123 wrong pt-BR zip ranges.
- **No locales at all** in gofakeit (#352, open) and Chance.
- **"Column-by-column" generation, no relationships between columns/tables** as a general critique of Faker-style tools (https://securityboulevard.com/2026/04/best-synthetic-data-generation-tools-and-platforms-compared-for-2026/).
- **Uniqueness**: faker-js removed `faker.unique` and tells users to roll their own (https://fakerjs.dev/guide/unique.html); Python `fake.unique` raises `UniquenessException` after retries.

## 4. Notable and unusual features

- **Datafaker** — expression language `#{Name.first_name}`, `#{regexify '[a-z]{4,10}'}`, `#{numerify '##'}`, `#{bothify}`, `#{letterify}`, `#{templatify}`, `#{examplify}`, `#{options.option 'A','B'}`, `#{csv …}`, `#{json …}`, method calls with args (`#{date.birthday 'yy DDD'}`); nested resolution (https://www.datafaker.net/documentation/expressions/). `Schema.of(field(...), compositeField(...))` with transformers to CSV, JSON, YAML, XML, TOML, SQL (batch inserts, Postgres/Oracle/Spark dialects, ARRAY/MULTISET/ROW), `@FakeForSchema` POJO population (https://www.datafaker.net/documentation/schemas/). `faker.collection(...).len(3,5)`, infinite `stream()`, `faker.unique()`, sequences, custom yml via `addPath`. 263 providers, the largest pop-culture catalogue (≈110 franchises/games).
- **gofakeit** — Go `text/template` based `Template()`, `Markdown()`, `EmailText()`, `FixedWidth()`; `Generate("{firstname} {regex:[a-z]{5}} {randomstring:[a,b]}")` mini-language; `Struct(&v)` fill via `fake:"{city}"`, `fakesize`, `format` tags; `Slice`, `Map`, `Regex`; `CSV/JSON/XML/SQL` output; `AddFuncLookup` registers custom functions with parameter metadata (used by its CLI and HTTP server); `Weighted()`; word families split by grammatical role (46 Noun/Verb/Adjective/Pronoun/Connective sub-kinds); `Error*` generators; Minecraft.
- **Mimesis** — `Field`/`Fieldset`/`Schema` builder (`Field(locale)("person.full_name")`, keyed lookups with `key=` post-processors like `maybe`, `romanize`), `Schema(schema=lambda: {...}, iterations=n)` exporting to JSON/CSV/pickle, relational data with `SchemaRef` foreign keys, custom field handlers, factory-boy integration, fastest Python generator; locale = six JSON files; ISO country codes in A2/A3/numeric; DMS coordinates.
- **Bogus** — `Faker<T>().RuleFor(x => x.Prop, f => …)` fluent rules with `StrictMode`, `CustomInstantiator`, `FinishWith`, `Rules()`, `Ignore`, rule sets; local vs global seeding, `UseDateTimeReference`; `Bogus.Distributions.Gaussian`; country ID extension packages; premium Locations/Healthcare/Hollywood/Text packages; Roslyn analyzer; `AutoBogus` auto-fills any class.
- **faker-js** — locale fallback stack with explicit `null` = "not applicable"; `mergeLocales`; `helpers.fake('{{person.firstName}}')` template; `helpers.fromRegExp`, `weightedArrayElement`, `multiple`, `uniqueArray`, `mustache`; pluggable `Randomizer` (32/53-bit Mersenne), `Distributors` (uniform, exponential); `date.birthdate({mode:'age'|'year'})`; `location.nearbyGPSCoordinate`; `git` module; `person.bio`; typed per-locale definition files generated by script.
- **Python Faker** — `use_weighting` (frequency-weighted names/formats from census data); `OrderedDict` weighted `random_element`; `fake.unique`, `fake.optional`; multi-locale instance with weights; `geo.local_latlng` real geonames records; `misc` file/archive/CSV/JSON/image generators; `pystr_format`; pytest fixture; CLI; 24 community providers (healthcare, market data, observability, airtravel, education, security, pyspark …).
- **Chance** — `weighted`, `normal` (Gaussian), `unique`, `n`, `mixin`, `set` (override data), `pick/pickone/pickset`; mobile IDs (android_id, apple_token, bb_pin, wp7/8 anid); geohash/altitude/depth.
- **Ruby faker** — `full_address_as_hash`, `Faker::Config.locale`, `Faker::UniqueGenerator`, `Faker::Config.random`; `Types` generator for random Ruby types; company registration numbers for ~12 countries; deep pop-culture catalogue (~85 themed generators).
