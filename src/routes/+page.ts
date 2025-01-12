import { fetchCoupons, fetchProducts } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	return {
		items: await fetchProducts(),
		coupons: await fetchCoupons(),
	};
};
