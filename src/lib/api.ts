import { z, type ZodType } from 'zod'
import { AdminCoupon, Coupon, OrderItem, type CartItem } from './cart'
import { ShopItem } from './shop'
import { browser } from '$app/environment'

const API_URL = `${import.meta.env.VITE_URL}/api/v0`
const APIStoreClosureError = z.object({
	type: z.literal('store_closure'),
	end_time: z.number().nullable(),
	message: z.string(),
	show_order_check: z.boolean()
})

export class StoreClosureError extends Error {
	end_time: number | null
	user_message: string
	allow_order: boolean

	constructor(end_time: number | null, message: string, allow_order: boolean) {
		super(`The store is currently closed. Admin reason: ${message}`)
		this.name = 'StoreClosureError'
		this.end_time = end_time
		this.user_message = message
		this.allow_order = allow_order
	}
}

let fetch = browser
	? window.fetch
	: () => {
			return { json: () => {}, text: () => '', status: 404 }
		}

export const setFetch = (newFetch: typeof fetch) => {
	fetch = newFetch
}

const handleFetch = async <T extends ZodType>(
	typ: T,
	path: string,
	body?: Object,
	params?: RequestInit
): Promise<z.infer<T>> => {
	const resp = await fetch(`${API_URL}${path}`, {
		method: body ? 'POST' : 'GET',
		body: body ? JSON.stringify(body) : undefined,
		...params
	})
	switch (resp.status) {
		case 401:
			window.location.replace(`${API_URL}/auth`)
			throw new Error('Pending authentication')
		case 200:
			break
		case 204:
			return
		default:
			const val = await resp.text()
			throw new Error(`unexpected status code ${resp.status}: ${val}`)
	}
	const val = await resp.json()
	const parseResult = typ.safeParse(val)
	if (!parseResult.success) {
		const err = APIStoreClosureError.safeParse(val)
		if (err.success) {
			throw new StoreClosureError(err.data.end_time, err.data.message, err.data.show_order_check)
		}
		throw parseResult.error
	}
	return parseResult.data
}

const orders = {
	checkout: async (
		checkoutItems: CartItem[],
		name: string,
		matricNumber: string,
		email: string,
		coupon: string | undefined
	): Promise<string> => {
		const items = checkoutItems.map((x) => ({
			id: x.id,
			variant: x.variant,
			amount: x.amount
		}))
		const resp = await handleFetch(z.object({ checkoutURL: z.string() }), `/checkout`, {
			name,
			matricNumber,
			email,
			coupon,
			items
		})
		return resp.checkoutURL
	},
	search: async (keyword: string): Promise<Order[]> => {
		let url = `/orders/${encodeURI(keyword)}`
		const resp = await handleFetch(z.object({ orders: Order.array() }), url)
		return resp.orders
	}
} as const

const sales = (period: string = 'current') =>
	({
		coupons: async (): Promise<Coupon[]> => {
			const resp = await handleFetch(
				z.object({ coupons: Coupon.array() }),
				`/sales/${period}/coupons`
			)
			return resp.coupons
		},
		couponByCode: async (code: string): Promise<Coupon> =>
			handleFetch(Coupon, `/sales/${period}/coupons/${encodeURI(code)}`),
		products: async (): Promise<ShopItem[]> => {
			const resp = await handleFetch(
				z.object({ products: ShopItem.array() }),
				`/sales/${period}/products`
			)
			return resp.products
		}
	}) as const

const adminSales = (period: string = 'current') =>
	({
		coupons: async (): Promise<AdminCoupon[]> => {
			const resp = await handleFetch(
				z.object({ coupons: AdminCoupon.array() }),
				`/sales/${period}/coupons?include_disabled=1`
			)
			return resp.coupons
		},
		products: async (): Promise<ShopItem[]> => {
			const resp = await handleFetch(
				z.object({ products: ShopItem.array() }),
				`/sales/${period}/products?include_disabled=1`
			)
			return resp.products
		},

		updateCoupon: async (coupon: AdminCoupon): Promise<AdminCoupon> =>
			handleFetch(AdminCoupon, `/sales/${period}/coupons`, coupon),
		updateProduct: async (product: ShopItem): Promise<ShopItem> =>
			handleFetch(ShopItem, `/sales/${period}/products`, product),

		orderSummary: async (showCollected: boolean): Promise<OrderSummary> => {
			let url = `/sales/${period}/order_summary`
			if (showCollected) url += '?show_collected=1'
			return handleFetch(OrderSummary, url)
		}
	}) as const

const adminOrders = {
	cancel: (orderID: string): Promise<void> =>
		handleFetch(z.undefined(), `/orders/${encodeURI(orderID)}/cancel`, {}),
	collect: (orderID: string): Promise<void> =>
		handleFetch(z.undefined(), `/orders/${encodeURI(orderID)}/collect`, {}),
	search: async (keyword: string, includeCancelled?: boolean): Promise<Order[]> => {
		let url = `/orders/${encodeURI(keyword)}?from_item=1`
		if (includeCancelled) url += '&include_cancelled=1'
		const resp = await handleFetch(z.object({ orders: Order.array() }), url)
		return resp.orders
	}
} as const

const adminClosures = {
	list: async (): Promise<StoreClosure[]> => {
		const resp = await handleFetch(z.object({ closures: StoreClosure.array() }), `/closures`)
		return resp.closures
	},
	update: (closure: StoreClosure): Promise<StoreClosure> =>
		handleFetch(StoreClosure, '/closures', closure)
}

export type User = z.infer<typeof User>
const User = z.string()
const adminUsers = {
	list: async (): Promise<User[]> => {
		const resp = await handleFetch(z.object({ users: User.array() }), '/users')
		return resp.users
	},
	add: (user: User): Promise<void> => handleFetch(z.undefined(), '/users', user),
	remove: (user: User): Promise<void> =>
		handleFetch(z.undefined(), '/users', user, { method: 'DELETE' })
} as const

const admin = {
	closures: adminClosures,
	orders: adminOrders,
	sales: adminSales,
	users: adminUsers,
	checkPerm: (): Promise<void> => handleFetch(z.undefined(), '/perm_check'),
	listSales: async (): Promise<SalePeriod[]> => {
		const resp = await handleFetch(z.object({ periods: SalePeriod.array() }), `/sales`)
		return resp.periods
	},
	updateSales: (period: SalePeriod): Promise<SalePeriod> =>
		handleFetch(SalePeriod, '/sales', period),
	uploadImage: async (img: Blob, raw: boolean = false): Promise<string> => {
		const data = new FormData()
		data.append('file', img)
		data.append('raw', `${raw ? 1 : 0}`)
		const resp = await handleFetch(z.object({ url: z.string() }), `/image_upload`, undefined, {
			method: 'POST',
			body: data
		})
		return resp.url
	}
} as const

const Order = z.object({
	id: z.string(),
	name: z.string(),
	matricNumber: z.string(),
	email: z.string(),
	paymentRef: z.string(),
	paymentTime: z.coerce.date().nullable(),
	collectionTime: z.coerce.date().nullable(),
	cancelled: z.boolean(),
	coupon: Coupon.nullable(),
	items: OrderItem.array(),
	salePeriod: z.string()
})
const StoreClosure = z.object({
	id: z.string(),
	start_time: z.coerce.date(),
	end_time: z.coerce.date(),
	message: z.string(),
	show_order_check: z.boolean()
})
const OrderSummary = z.object({
	unfulfilled: z
		.object({
			name: z.string(),
			variant: z.string(),
			count: z.number()
		})
		.array(),
	order_id_samples: z.string().array(),
	unfulfilled_order_count: z.number(),
	fulfilled_order_count: z.number()
})
const SalePeriod = z.object({
	id: z.string(),
	name: z.string(),
	start_time: z.coerce.date()
})

export type Order = z.infer<typeof Order>
export type StoreClosure = z.infer<typeof StoreClosure>
export type OrderSummary = z.infer<typeof OrderSummary>
export type SalePeriod = z.infer<typeof SalePeriod>

export default {
	admin,
	orders,
	sales
} as const
