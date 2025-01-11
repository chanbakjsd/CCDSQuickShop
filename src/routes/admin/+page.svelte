<script lang="ts">
	import { onMount } from 'svelte';

	import { permCheck } from '$lib/api';
	import Header from '$lib/Header.svelte';
	import Options from '$lib/Options.svelte';
	import MerchEdit from './MerchEdit.svelte';
	import UsersEdit from './UsersEdit.svelte';

	onMount(() => {
		permCheck();
	});

	let options = [{ text: 'Merch' }, { text: 'Admin Users' }, { text: 'Order Collection' }];
	let selected: string | undefined = undefined;
</script>

<div class="flex flex-col gap-4 p-4">
	<Header admin />
	<Options {options} bind:value={selected} />
	{#if selected === 'Merch'}
		<MerchEdit />
	{:else if selected === 'Admin Users'}
		<UsersEdit />
	{:else if selected === 'Order Collection'}
		TODO
	{/if}
</div>
