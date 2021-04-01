<script>
    export let lable;
    export let value;
</script>

<div>
    <lable for={lable}><b>{lable}</b></lable><br>
    <input name={lable} placeholder={lable} {...$$restProps} bind:value />
</div>
