import { z } from 'zod'

export const ShopItem = z.object({
	id: z.string(),
	name: z.string(),
	basePrice: z.number(),
	variants: z
		.object({
			type: z.string(),
			options: z
				.object({
					text: z.string(),
					additionalPrice: z.number().optional()
				})
				.array()
		})
		.array(),
	defaultImageURL: z.string(),
	imageURLs: z
		.object({
			selectedOptions: z.string().nullable().array(),
			url: z.string()
		})
		.array(),
	enabled: z.boolean().optional(),
	salePeriod: z.number()
})

export type ShopItem = z.infer<typeof ShopItem>

export const emptyShopItem = (name: string, salePeriod: string): ShopItem => ({
	id: '',
	name,
	basePrice: 0,
	variants: [],
	defaultImageURL: '',
	imageURLs: [],
	enabled: false,
	salePeriod: +salePeriod
})

export const toArrayVariant = (item: ShopItem, variants: Record<string, string>) => {
	const arrVarriants: (string | undefined)[] = Array(item.variants.length).fill(undefined)
	for (let i = 0; i < item.variants.length; i++) {
		const variant = item.variants[i]
		if (variant.type in variants) {
			arrVarriants[i] = variants[variant.type]
		}
	}
	return arrVarriants
}

export const resolveImageURL = (item: ShopItem, variants: Record<string, string>) => {
	const arrVariants = toArrayVariant(item, variants)
	// Find the first best match (undefined in candidates means don't care).
	let bestMatch = 0
	let result = item.defaultImageURL
	for (const candidate of item.imageURLs) {
		const foundMismatch = arrVariants.some(
			(x, i) => candidate.selectedOptions[i] && candidate.selectedOptions[i] !== x
		)
		if (foundMismatch) continue
		const matchCount = candidate.selectedOptions.filter((x) => !!x).length
		if (matchCount > bestMatch) {
			bestMatch = matchCount
			result = candidate.url
		}
	}
	return result
}

export const tentativePrice = (item: ShopItem, variants: Record<string, string>) => {
	const arrVariants = toArrayVariant(item, variants)
	// Add all the existing addon prices to the base price.
	let price = item.basePrice
	for (let i = 0; i < arrVariants.length; i++) {
		if (!arrVariants[i]) continue
		const addonCost = item.variants[i].options.find(
			(x) => x.text === arrVariants[i]
		)?.additionalPrice
		price += addonCost ?? 0
	}
	return price
}
