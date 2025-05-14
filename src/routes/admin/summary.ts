import type { OrderSummary } from '$lib/api'

type Table = {
	name: string
	columns: TableColumn[][]
	rows: TableRow[]
}

type TableColumn = {
	label: string
	span: number
}
type TableRow = {
	label: string
	fullLabels: string[]
	data: number[]
}

export const constructTables = (summary: OrderSummary): Table[] => {
	const variants = productVariants(summary)
	const productGroups = findProductGroups(variants)
	return productGroups.map((group) => constructTable(group, variants, summary))
}

type ProductName = string
type ProductVariant = string[][]
type SourceRowEntry = {
	product: string
	variant: string[]
	count: number
}

const productVariants = (summary: OrderSummary): Record<ProductName, ProductVariant> => {
	// Black, L and White, XL should generate [Black, White], [L, XL].
	const productVariants: Record<string, string[][]> = {}
	summary.unfulfilled.forEach((x) => {
		const variantEntries = x.variant.split(',').map((v) => v.trim())
		if (!(x.name in productVariants)) {
			productVariants[x.name] = variantEntries.map((v) => [v])
			return
		}
		variantEntries.forEach((y, i) => {
			if (productVariants[x.name].length <= i) {
				productVariants[x.name].push([y])
				return
			}
			if (productVariants[x.name][i].some((v) => v === y)) {
				// It already exists inside the correct array.
				return
			}
			productVariants[x.name][i].push(y)
		})
	})
	return productVariants
}

const findProductGroups = (
	productVariants: Record<ProductName, ProductVariant>
): ProductName[][] => {
	// Naively combine the products. It's better than nothing, especially
	// if there are multiple shop products that are actually just variants.
	const productGroups: ProductName[][] = []
	let ungrouped = Object.keys(productVariants)
	while (ungrouped.length > 0) {
		const leader = productVariants[ungrouped[0]]
		const newGroup = [ungrouped[0]]
		ungrouped.forEach((candidate, i) => {
			if (i === 0) return
			if (productVariants[candidate].length !== leader.length) {
				return
			}
			let candidateOK = true
			for (const [i, variant] of productVariants[candidate].entries()) {
				if (!variant.some((x) => leader[i].includes(x))) {
					// There isn't a variant that the leader has, that probably means it's very different.
					// e.g. Comparing White, Black with S, L
					candidateOK = false
					break
				}
			}
			if (candidateOK) {
				newGroup.push(candidate)
			}
		})
		productGroups.push(newGroup)
		ungrouped = ungrouped.filter((x) => !newGroup.includes(x))
	}
	return productGroups
}

const relevantEntriesForProducts = (
	summary: OrderSummary,
	products: ProductName[]
): SourceRowEntry[] =>
	summary.unfulfilled
		.filter((x) => products.includes(x.name))
		.map((x) => ({
			product: x.name,
			variant: x.variant.split(',').map((v) => v.trim()),
			count: x.count
		}))

// Assumption: Length is same for each entry, which is true from findProductGroups.
const combineVariants = (
	products: ProductName[],
	productVariants: Record<ProductName, ProductVariant>
): ProductVariant =>
	productVariants[products[0]].map((_, i) =>
		[...new Set(products.flatMap((x) => productVariants[x][i]))].sort(variantSort)
	)

const SORT_ORDER = ['3XS', '2XS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL']
const variantSort = (a: string, b: string) => {
	// Use SORT_ORDER if applicable to make sure size is in the right order.
	const aIndex = SORT_ORDER.indexOf(a)
	const bIndex = SORT_ORDER.indexOf(b)
	if (aIndex === -1 || bIndex === -1) return a < b ? -1 : a === b ? 0 : 1
	return aIndex - bIndex
}

const constructTable = (
	group: ProductName[],
	variants: Record<ProductName, ProductVariant>,
	summary: OrderSummary
): Table => {
	const name = group.join(', ')
	const variant = combineVariants(group, variants)
	const entries = relevantEntriesForProducts(summary, group)
	// Columns of [A,B], [1,2,3,4,5] will look like:
	//   A    B
	// 1234512345
	const { columnIdx, columns, columnCount } = columnsFromVariant(variant)
	const rowIdx = []
	for (let i = 0; i < variant.length; i++) {
		// Everything not involved in columnIdx should be part of row.
		if (columnIdx.includes(i)) continue
		rowIdx.push(i)
	}
	const rowsRecord: Record<string, TableRow> = {}
	for (const entry of entries) {
		let label = entry.product
		if (group.length === 1 && rowIdx.length > 0) {
			// There is only one product, no point labelling it.
			label = ''
		}
		for (const i of rowIdx) {
			if (label === '') {
				label = entry.variant[i]
				continue
			}
			label += ', ' + entry.variant[i]
		}
		if (!(label in rowsRecord)) {
			rowsRecord[label] = {
				label,
				data: new Array(columnCount).fill(0),
				fullLabels: new Array(columnCount).fill('')
			}
		}
		let columnLoc = 0
		let currentLevel = 1
		for (let columnIdxIdx = columnIdx.length - 1; columnIdxIdx >= 0; columnIdxIdx--) {
			const i = columnIdx[columnIdxIdx]
			// Counting except every single digit has a different base.
			// currentLevel is the base of the current variant.
			columnLoc += variant[i].indexOf(entry.variant[i]) * currentLevel
			currentLevel *= variant[i].length
		}
		rowsRecord[label].fullLabels[columnLoc] = [entry.product, ...entry.variant].join(', ')
		rowsRecord[label].data[columnLoc] += entry.count
	}
	const rows = Object.keys(rowsRecord)
		.map((k) => rowsRecord[k])
		.sort((a, b) => {
			const aLabel = a.label.split(', ')
			const bLabel = b.label.split(', ')
			for (let i = 0; i < Math.min(aLabel.length, bLabel.length); i++) {
				const sortRes = variantSort(aLabel[i], bLabel[i])
				if (sortRes !== 0) return sortRes
			}
			return variantSort(a.label, b.label)
		})
	return { name, columns, rows }
}

const MAX_COLUMNS = 12
const columnsFromVariant = (variant: ProductVariant) => {
	let totalColumnCount = 1
	let columnEntries = []
	for (let i = 0; i < variant.length; i++) {
		if (totalColumnCount * variant[i].length <= MAX_COLUMNS) {
			columnEntries.push(i)
			totalColumnCount *= variant[i].length
		}
	}
	const result: TableColumn[][] = []
	let repeatCount = 1
	let remainingColumnCount = totalColumnCount
	for (const i of columnEntries) {
		const rowOfCol: TableColumn[] = []
		remainingColumnCount /= variant[i].length
		for (let j = 0; j < repeatCount; j++) {
			for (const v of variant[i]) {
				rowOfCol.push({
					label: v,
					span: remainingColumnCount
				})
			}
		}
		repeatCount *= variant[i].length
		result.push(rowOfCol)
	}
	return {
		columnIdx: columnEntries,
		columns: result,
		columnCount: totalColumnCount
	}
}
