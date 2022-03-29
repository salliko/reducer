package databases

var (
	createTable = `
		create table if not exists urls (
			id serial primary key not null,
			hash varchar(25),
			original text,
			user_id varchar(250),
			is_deleted boolean default false
		)
	`

	insert = `
		insert into urls (hash, original, user_id) 
		values ($1, $2, $3)
	`

	selectOriginal = `
		select original, is_deleted from urls where hash = $1
	`

	selectAllUserRows = `
		select 
			hash, original, user_id
		from urls
		where user_id = $1
	`

	delete = `
		update urls set
			is_deleted = true
		where hash = $1 and user_id = $2
	`
)
