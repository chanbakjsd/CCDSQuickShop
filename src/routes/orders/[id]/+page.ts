import { listOrders, type Order } from '$lib/api'
import { error } from '@sveltejs/kit'
import type { PageLoad } from './$types'
import { sortOrder } from '$lib/util'

export const load: PageLoad = async ({ params }) => {
	let orders: Order[]
	try {
		orders = await listOrders(params.id)
	} catch (e) {
		console.warn('Error: ', e)
		error(500, 'Your order cannot be checked at this time.')
	}
	if (orders.length === 0) {
		error(404, 'The specified order does not exist.')
	}
	return { orders: sortOrder(orders) }
}
