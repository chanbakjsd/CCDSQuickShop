import { z, type ZodType } from 'zod'
import { AdminCoupon, Coupon, OrderItem, type CartItem } from './cart'
import { ShopItem } from './shop'

const API_URL = `${import.meta.env.VITE_URL}/api/v0`

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

const handleFetch = async <T extends ZodType>(
	typ: T,
	path: string,
	params?: RequestInit
): Promise<z.infer<T>> => {
	const resp = await fetch(path, params)
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
	if (parseResult.success) {
		return parseResult.data
	}
	const err = APIStoreClosureError.safeParse(val)
	if (err.success) {
		throw new StoreClosureError(err.data.end_time, err.data.message, err.data.show_order_check)
	}
	throw parseResult.error
}

const APIStoreClosureError = z.object({
	type: z.literal('store_closure'),
	end_time: z.number().nullable(),
	message: z.string(),
	show_order_check: z.boolean()
})

const CheckoutResponse = z.object({
	checkoutURL: z.string()
})
export const checkout = async (
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
	const resp = await handleFetch(CheckoutResponse, `${API_URL}/checkout`, {
		method: 'POST',
		body: JSON.stringify({ name, matricNumber, email, coupon, items })
	})
	return resp.checkoutURL
}

const ProductsResponse = z.object({
	products: ShopItem.array()
})
export const fetchProducts = async (includeDisabled?: boolean): Promise<ShopItem[]> => {
	let path = `${API_URL}/products`
	if (includeDisabled) {
		path += '?include_disabled=1'
	}
	const resp = await handleFetch(ProductsResponse, path)
	return resp.products
}

export const updateProduct = async (product: ShopItem): Promise<ShopItem> => {
	return handleFetch(ShopItem, `${API_URL}/products`, {
		method: 'POST',
		body: JSON.stringify(product)
	})
}

export const fetchCoupon = (couponCode: string): Promise<Coupon> =>
	handleFetch(Coupon, `${API_URL}/coupons/${encodeURI(couponCode)}`)

const CouponsResponse = z.object({
	coupons: Coupon.array()
})
export const fetchCoupons = async (): Promise<Coupon[]> => {
	const resp = await handleFetch(CouponsResponse, `${API_URL}/coupons`)
	return resp.coupons
}

const AdminCouponsResponse = z.object({
	coupons: AdminCoupon.array()
})
export const fetchAdminCoupons = async (): Promise<AdminCoupon[]> => {
	const resp = await handleFetch(AdminCouponsResponse, `${API_URL}/coupons?include_disabled=1`)
	return resp.coupons
}

export const updateCoupon = async (coupon: AdminCoupon): Promise<AdminCoupon> =>
	handleFetch(AdminCoupon, `${API_URL}/coupons`, {
		method: 'POST',
		body: JSON.stringify(coupon)
	})

export const permCheck = async (): Promise<void> => {
	return handleFetch(z.undefined(), `${API_URL}/perm_check`)
}

export type User = z.infer<typeof User>
const User = z.string()
const UserResponse = z.object({
	users: User.array()
})

export const listUsers = async (): Promise<User[]> => {
	const resp = await handleFetch(UserResponse, `${API_URL}/users`)
	return resp.users
}

export const addUser = (user: User): Promise<void> =>
	handleFetch(z.undefined(), `${API_URL}/users`, {
		method: 'POST',
		body: JSON.stringify(user)
	})

export const deleteUser = (user: User): Promise<void> =>
	handleFetch(z.undefined(), `${API_URL}/users`, {
		method: 'DELETE',
		body: JSON.stringify(user)
	})

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
	items: OrderItem.array()
})
const OrdersResponse = z.object({
	orders: Order.array()
})
export type Order = z.infer<typeof Order>
export const listOrders = async (
	keyword: string,
	options?: { includeCancelled?: boolean; allowFromItem?: boolean }
): Promise<Order[]> => {
	let url = `${API_URL}/orders/${encodeURI(keyword)}`
	const opts = []
	if (options?.includeCancelled) opts.push('include_cancelled=1')
	if (options?.allowFromItem) opts.push('from_item=1')
	if (opts.length > 0) url += `?${opts.join('&')}`
	const resp = await handleFetch(OrdersResponse, url)
	return resp.orders
}

export const collectOrder = (orderID: string): Promise<void> =>
	handleFetch(z.undefined(), `${API_URL}/orders/${encodeURI(orderID)}/collect`, {
		method: 'POST'
	})
export const cancelOrder = (orderID: string): Promise<void> =>
	handleFetch(z.undefined(), `${API_URL}/orders/${encodeURI(orderID)}/cancel`, {
		method: 'POST'
	})

const ImageUploadResponse = z.object({
	url: z.string()
})
export const uploadImage = async (img: Blob): Promise<string> => {
	const data = new FormData()
	data.append('file', img)
	const resp = await handleFetch(ImageUploadResponse, `${API_URL}/image_upload`, {
		method: 'POST',
		body: data
	})
	return resp.url
}

const StoreClosure = z.object({
	id: z.string(),
	start_time: z.coerce.date(),
	end_time: z.coerce.date(),
	message: z.string(),
	show_order_check: z.boolean()
})
export type StoreClosure = z.infer<typeof StoreClosure>

const StoreClosureResponse = z.object({
	closures: StoreClosure.array()
})
export const listStoreClosures = async (): Promise<StoreClosure[]> => {
	const resp = await handleFetch(StoreClosureResponse, `${API_URL}/closures`)
	return resp.closures
}

export const updateStoreClosure = async (closure: StoreClosure): Promise<StoreClosure> => {
	return handleFetch(StoreClosure, `${API_URL}/closures`, {
		method: 'POST',
		body: JSON.stringify(closure)
	})
}

const UnfulfilledOrderSummary = z.object({
	unfulfilled: z
		.object({
			name: z.string(),
			variant: z.string(),
			count: z.number()
		})
		.array(),
	order_id_samples: z.string().array()
})
export type UnfulfilledOrderSummary = z.infer<typeof UnfulfilledOrderSummary>
export const unfulfilledOrderSummary = (): Promise<UnfulfilledOrderSummary> =>
	handleFetch(UnfulfilledOrderSummary, `${API_URL}/order_summary`)
