import { error, redirect } from '@sveltejs/kit';
import { StoreClosureError, fetchCoupons, fetchProducts } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	try {
		return {
			items: await fetchProducts(),
			coupons: await fetchCoupons(),
		};
	} catch (e) {
		if (e instanceof StoreClosureError) {
			redirect(307, "/wait")
		}
		error(500)
	}
};
