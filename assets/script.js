update_go_package = function(go_package_id) {
	var go_package = document.getElementById(go_package_id);
	var go_package_button = go_package.getElementsByClassName("update-button")[0];

	go_package_button.textContent = "Updating...";
	go_package_button.onclick = "javascript:void(0)"
	go_package_button.tabIndex = -1;
	go_package_button.className += " disabled";

	var request = new XMLHttpRequest();
	request.onload = function() {
		// Hide the "Updating..." label.
		go_package_button.style.display = "none";

		// Show "No Updates Available" if there are no remaining updates.
		var updates_available = 0;
		var updates = document.getElementsByClassName("go-package-update");
		for (var i = 0; i < updates.length; i++) {
			if (updates[i].getElementsByClassName("disabled").length == 0) {
				updates_available++;
				break;
			}
		}
		if (updates_available == 0) {
			document.getElementById("no_updates").style.display = "";
		}

		// Move this Go package to "Installed Updates" list.
		var installed_updates = document.getElementById("installed_updates");
		installed_updates.style.display = "";
		installed_updates.parentNode.insertBefore(go_package, installed_updates.nextSibling); // Insert after.
	};
	request.open('POST', '/-/update', true);
	request.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
	request.send("import_path_pattern=" + go_package_id);
}
