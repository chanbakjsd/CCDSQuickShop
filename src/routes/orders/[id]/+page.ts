import { listOrders } from '$lib/api'
import { error } from '@sveltejs/kit'
import type { PageLoad } from './$types'
import { sortOrder } from '$lib/util'

export const load: PageLoad = async ({ params }) => {
	try {
		const orders = await listOrders(params.id)
		if (orders.length === 0) {
			error(404, 'The specified order ID does not exist.')
		}
		return { orders: sortOrder(orders) }
	} catch (e) {
		console.warn('Error: ', e)
		error(500, 'Your order cannot be checked at this time.')
	}
}
