import { listOrders } from '$lib/api';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const orders = await listOrders(params.id)
	if (orders.length === 0) {
		error(404)
	}
	return { orders }
}
