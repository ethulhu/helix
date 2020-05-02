export async function fetchDirectories() {
	const rsp = await fetch( '/directories/', {
		headers: { Accept: 'application/json' },
	} );
	return rsp.json()
}

export async function fetchObject( directory, id ) {
	const rsp = await fetch( `/directories/${directory}/${id}`, {
		headers: { Accept: 'application/json' },
	} );
	return rsp.json()
}

export const rootObject = '0';
