update_go_package = function(go_package_button) {
	go_package_button.textContent = "Updating...";
	go_package_button.onclick = "javascript:void(0)"
	go_package_button.tabIndex = -1;
	go_package_button.className = "disabled";
	request = new XMLHttpRequest;
	request.open('POST', 'http://localhost:7043/-/update', true);
	request.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
	request.send("import_path_pattern=" + go_package_button.id);
}
