import { error, redirect } from '@sveltejs/kit';
import { StoreClosureError, fetchProducts } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	try {
		await fetchProducts()
	} catch (e) {
		if (e instanceof StoreClosureError) {
			return {
				end_time: e.end_time,
				user_message: e.user_message,
				allow_order: e.allow_order,
			}
		}
		error(500)
	}
	// If we succeed, just redirect back to store page.
	redirect(302, "/")
};
