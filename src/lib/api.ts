import { z, type ZodType } from "zod"
import { Coupon, OrderItem, type CartItem } from "./cart"
import { ShopItem } from "./shop"

const API_URL = `${import.meta.env.VITE_URL}/api/v0`

const handleFetch = async <T extends ZodType>(typ: T, path: string, params?: RequestInit): Promise<z.infer<T>> => {
	const resp = await fetch(path, params)
	switch (resp.status) {
		case 401:
			window.location.replace(`${API_URL}/auth`)
			throw new Error("Pending authentication")
		case 200:
			break
		case 204:
			return
		default:
			const val = await resp.text()
			throw new Error(`unexpected status code ${resp.status}: ${val}`)
	}
	const val = await resp.json()
	return typ.parse(val)
}

const CheckoutResponse = z.object({
	checkoutURL: z.string(),
})
export const checkout = async (checkoutItems: CartItem[], name: string, matricNumber: string, email: string, coupon: string | undefined): Promise<string> => {
	const items = checkoutItems.map(x => ({
		id: x.id,
		variant: x.variant,
		amount: x.amount,
	}))
	const resp = await handleFetch(CheckoutResponse, `${API_URL}/checkout`, {
		method: "POST",
		body: JSON.stringify({ name, matricNumber, email, coupon, items })
	})
	return resp.checkoutURL
}

const ProductsResponse = z.object({
	products: ShopItem.array(),
})
export const fetchProducts = async (includeDisabled?: boolean): Promise<ShopItem[]> => {
	let path = `${API_URL}/products`
	if (includeDisabled) {
		path += "?include_disabled=1"
	}
	const resp = await handleFetch(ProductsResponse, path)
	return resp.products
}

export const updateProduct = async (product: ShopItem): Promise<ShopItem> => {
	return handleFetch(ShopItem, `${API_URL}/products`, {
		method: 'POST',
		body: JSON.stringify(product),
	})
}

export const fetchCoupon = (couponCode: string): Promise<Coupon> => handleFetch(Coupon, `${API_URL}/coupons/${encodeURI(couponCode)}`)

const CouponsResponse = z.object({
	coupons: Coupon.array(),
})
export const fetchCoupons = async (includeDisabled?: boolean): Promise<Coupon[]> => {
	let path = `${API_URL}/coupons`
	if (includeDisabled) {
		path += "?include_disabled=1"
	}
	const resp = await handleFetch(CouponsResponse, path)
	return resp.coupons
}

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

export const addUser = (user: User): Promise<void> => handleFetch(z.undefined(), `${API_URL}/users`, {
	method: "POST",
	body: JSON.stringify(user),
})

export const deleteUser = (user: User): Promise<void> => handleFetch(z.undefined(), `${API_URL}/users`, {
	method: "DELETE",
	body: JSON.stringify(user),
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
	items: OrderItem.array(),
})
const OrdersResponse = z.object({
	orders: Order.array()
})
export type Order = z.infer<typeof Order>;
export const listOrders = async (keyword: string, includeCancelled?: boolean): Promise<Order[]> => {
	let url = `${API_URL}/orders/${encodeURI(keyword)}`
	if (includeCancelled) {
		url += "?include_cancelled=1"
	}
	const resp = await handleFetch(OrdersResponse, url)
	return resp.orders
}

export const collectOrder = (orderID: string): Promise<void> => handleFetch(z.undefined(), `${API_URL}/orders/${encodeURI(orderID)}/collect`, {
	method: "POST",
})
export const cancelOrder = (orderID: string): Promise<void> => handleFetch(z.undefined(), `${API_URL}/orders/${encodeURI(orderID)}/cancel`, {
	method: "POST",
})
