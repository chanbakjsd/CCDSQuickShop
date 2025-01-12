<script lang="ts">
	import { cancelOrder, collectOrder, listOrders, type Order } from '$lib/api';
	import Button from '$lib/Button.svelte';
	import Input from '$lib/Input.svelte';
	import OrderPreview from '../orders/[id]/OrderPreview.svelte';

	let orders: Order[] = [];
	let orderInput = '';
	const searchOrders = async () => {
		orders = await listOrders(orderInput);
	};
	const markCollect = (orderID: string) => async () => {
		await collectOrder(orderID);
		orders = orders.map((x) => ({
			...x,
			collectionTime: x.id === orderID ? new Date() : x.collectionTime
		}));
	};
	const markCancel = (orderID: string) => async () => {
		await cancelOrder(orderID);
		orders = orders.map((x) => ({
			...x,
			cancelled: x.id === orderID ? true : x.cancelled
		}));
	};
</script>

<div class="flex gap-2">
	<form on:submit={searchOrders}>
		<Input label="Matric Number/Order ID/NTU Email" bind:value={orderInput} />
	</form>
	<Button size="md" on:click={searchOrders}>Search</Button>
</div>
<div class="flex flex-col gap-4">
	{#each orders as order}
		<OrderPreview {order} collect={markCollect(order.id)} cancel={markCancel(order.id)} />
	{/each}
</div>
