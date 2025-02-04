const padZero = (s: number) => '0'.repeat(2 - (s + '').length) + s;
export const formatDate = (date: Date | null, connector = ' ') =>
	date
		? `${date.getFullYear()}-${padZero(date.getMonth() + 1)}-${padZero(date.getDate())}${connector}${padZero(date.getHours())}:${padZero(date.getMinutes())}:${padZero(date.getSeconds())}`
		: 'N/A';
