import { error, redirect } from '@sveltejs/kit'
import api, { StoreClosureError } from '$lib/api'
import type { PageLoad } from './$types'

export const load: PageLoad = async () => {
	try {
		await api.sales().products()
	} catch (e) {
		if (e instanceof StoreClosureError) {
			return {
				end_time: e.end_time,
				user_message: e.user_message,
				allow_order: e.allow_order
			}
		}
		if (e instanceof Error || typeof e === 'string') {
			error(500, e)
		}
		error(500)
	}
	// If we succeed, just redirect back to store page.
	redirect(302, '/')
}
