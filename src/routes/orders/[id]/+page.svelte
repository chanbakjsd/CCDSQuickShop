<script lang="ts">
	import Header from '$lib/Header.svelte';
	import type { PageData } from './$types';
	import OrderPreview from './OrderPreview.svelte';

	const { data }: { data: PageData } = $props();
	const { orders } = data;
</script>

<svelte:head>
	<title>Orders - SCDS Merch Store</title>
</svelte:head>

<div class="min-w-screen flex min-h-screen flex-col items-center gap-4 bg-gray-100 p-4">
	<Header cls="bg-gray-100" />
	{#each orders as order}
		<OrderPreview {order} expanded={orders.length === 1} userFacing />
	{/each}
</div>
