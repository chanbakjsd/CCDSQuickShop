import type { Order } from './api'

const padZero = (s: number) => '0'.repeat(2 - (s + '').length) + s

export const formatDate = (date: Date | null, connector = ' ') => {
	if (!date) return 'N/A'
	const d = `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())}`
	const time = `${padZero(date.getHours())}:${padZero(date.getMinutes())}:${padZero(date.getSeconds())}`
	return `${d}${connector}${time}`
}

export const sortOrder = (orders: Order[]) =>
	orders.sort((a, b) => {
		if (a.collectionTime != b.collectionTime) {
			if (a.collectionTime === null) return -1
			if (b.collectionTime === null) return 1
			return a.collectionTime > b.collectionTime ? -1 : 1
		}
		if (a.salePeriod != b.salePeriod) {
			return a.salePeriod < b.salePeriod ? 1 : -1
		}
		return a.id < b.id ? -1 : 1
	})
