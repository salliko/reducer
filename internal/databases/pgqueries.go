package databases

var (
	createTable = `
		create table if not exists urls (
			id serial primary key not null,
			hash varchar(25),
			original text,
			user_id varchar(250)
		)
	`

	insert = `
		insert into urls (hash, original, user_id) 
		values ($1, $2, $3)
	`

	hasValue = `
		select original from urls where hash = $1
	`

	selectOriginal = `
		select original from urls where hash = $1
	`

	selectAllUserRows = `
		select 
			hash, original, user_id
		from urls
		where user_id = $1
	`
)
