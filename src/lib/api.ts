import { z, type ZodType } from "zod"
import { ShopItem } from "./shop"

const handleFetch = async <T extends ZodType>(typ: T, path: string, params?: RequestInit): Promise<z.infer<T>> => {
	const resp = await fetch(path, params)
	switch (resp.status) {
		case 401:
			window.location.replace("/api/v0/auth")
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

const ProductsResponse = z.object({
	products: ShopItem.array(),
})
export const fetchProducts = async (includeDisabled?: boolean): Promise<ShopItem[]> => {
	let path = "/api/v0/products"
	if (includeDisabled) {
		path += "?include_disabled=1"
	}
	const resp = await handleFetch(ProductsResponse, path)
	return resp.products
}

export const updateProduct = async (product: ShopItem): Promise<ShopItem> => {
	return handleFetch(ShopItem, "/api/v0/products", {
		method: 'POST',
		body: JSON.stringify(product),
	})
}

export const permCheck = async (): Promise<void> => {
	return handleFetch(z.undefined(), "/api/v0/perm_check")
}

export type User = z.infer<typeof User>
const User = z.string()
const UserResponse = z.object({
	users: User.array()
})

export const listUsers = async (): Promise<User[]> => {
	const resp = await handleFetch(UserResponse, "/api/v0/users")
	return resp.users
}

export const addUser = (user: User): Promise<void> => handleFetch(z.undefined(), "/api/v0/users", {
	method: "POST",
	body: JSON.stringify(user),
})

export const deleteUser = (user: User): Promise<void> => handleFetch(z.undefined(), "/api/v0/users", {
	method: "DELETE",
	body: JSON.stringify(user),
})
