import { error, redirect } from '@sveltejs/kit'
import api, { setFetch, StoreClosureError } from '$lib/api'
import type { PageLoad } from './$types'

export const load: PageLoad = async ({ fetch }) => {
	setFetch(fetch)
	try {
		return {
			items: await api.sales().products(),
			coupons: await api.sales().coupons()
		}
	} catch (e) {
		if (e instanceof StoreClosureError) {
			redirect(307, '/wait')
		}
		if (e instanceof Error || typeof e === 'string') {
			error(500, e)
		}
		error(500)
	}
}
