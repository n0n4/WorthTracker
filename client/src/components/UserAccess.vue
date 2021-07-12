<template>
	<div>
		<h4>WorthTracker</h4>
        <!--User Controls-->
        <div class="card-footer">
            <form @submit.prevent="adduser">
                <div class="input-group">
                    <input class="form-control" placeholder="New User" id="adduser" type="text" v-model="newusername" autocomplete="off" required>
                    <span class="input-group-btn">
                        <button class="btn btn-secondary" type="submit">Add User</button>
                    </span>
                </div>
            </form>
        </div>
        <div class="card-footer">
            <form id = "selectuserform" @submit.prevent="selectuser">
                <select v-model="usertarget" required>
                    <option disabled value="">Select a User</option>
                    <option v-for="(user,userindex) in usernames" :key="'user' + userindex">{{ user }}</option>
                </select>
                <span class="input-group-btn">
                    <button class="btn btn-secondary" type="submit">View User</button>
                </span>
            </form>
        </div>
        <!--Statistic Display-->
        <span v-if="userselected"> 
            Net Worth: {{ this.networth }} <br/>
            Total Assets: {{ this.totalassets }} <br/>
            Total Liabilities: {{ this.totalliabilities }} <br/>
        </span>
        <!--Item Display-->
        <span v-if="userselected"> 
            <div class="container">
                <div class="card itemlistcard">
                    <div class="card-header itemlistheader">
                        Items
                    </div>
                    <!--New Item-->
                    <div class="card">
                        <form @submit.prevent="additem">
                            <div class="input-group">
                                <input class="form-control" placeholder="New Item" id="additemname" type="text" v-model="additemname" autocomplete="off" required>
                                <select v-model="additemtype" required>
                                    <option disabled value="">Select an Item Type</option>
                                    <option>Asset</option>
                                    <option>Liability</option>
                                </select>
                                <input class="form-control" placeholder="Value" id="additemvalue" type="text" v-model="additemvalue" autocomplete="off" required>
                                <span class="input-group-btn">
                                    <button class="btn btn-secondary" type="submit">Add Item</button>
                                </span>
                            </div>
                        </form>
                    </div>
                    <!--Item List-->
                    <table class="card table table-striped">
                        <thead>
                            <tr>
                                <th>Name</th>
                                <th>Type</th>
                                <th>Value</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="(item,index) in items" :key="'item' + itemcrutch + 'k' + index">
                                <td>{{ item.Name }}</td>  
                                <td>{{ item.Type }}</td> 
                                <td>{{ item.Value }}</td> 
                                <td>
                                    <button class="btn btn-secondary" @click="deleteitem(item.Id)">Delete</button>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </span>
        <span v-else> 
            Please select a user to view their account.
        </span>
	</div>
</template>

<style>
.card-header {
	text-align: left;
}
.card-footer {
	text-align: left;
}
.itemlistheader {
	font-weight: bold;
}
.itemlistcard {
	margin-top: 10px;
}
</style>

<script>
import axios from 'axios'

const axiosConfig = {
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
};

export default {
	data() {
		return {
			users: [], 	   // {Name: "", Id: #}
            usernames: [], // ["", "", "", ...]
            usercrutch: 0, // vue doesn't know to update the user list unless this key is updated
            // add user values
            newusername: "",
            // whether a user has been selected
            userselected: false,
            // overall statistics
            networth: 0,
            totalassets: 0,
            totalliabilities: 0,
            // add item values
            additemname: "",
            additemvalue: 0,
            additemtype: "",
            // item information
            items: [],
            itemcrutch: 0,
		}
	},
    created() {
		// load list of users
        this.refreshUsers();
	},
	methods: {
        refreshUsers: function() {
            axios.get('http://' + this.$addr + '/api/user',
                {},
                axiosConfig)
            .then(response => {
                // replace the users list
                this.users = response.data;
                this.makeUserNames();
                this.usercrutch++;
            })
            .catch(error => {
                alert("Failed to get users: " + error);
            });
        },
        refreshItems: function() {
            axios.post(
                'http://' + this.$addr + '/api/itemlist',
                { Username: this.usertarget },
                axiosConfig)
            .then(response => {
                // replace the items list
                this.items = response.data.Items;
                this.itemcrutch++;

                // loop over the items and divide by 100
                for(var i = 0; i < this.items.length; i++) {
                    this.items[i].Value /= 100;
                }

                this.networth = response.data.NetWorth / 100;
                this.totalassets = response.data.AssetTotal / 100;
                this.totalliabilities = response.data.LiabilityTotal / 100;

                console.log(JSON.stringify(this.items))
            })
            .catch(error => {
                alert("Failed to get items: " + error);
            });
        },
		makeUserNames: function() {
			let ns = [];
			for(var i = 0; i < this.users.length; i++) {
				ns.push(this.users[i].Name);
			}
			this.usernames = ns;
            console.log("Users: " + JSON.stringify(this.users));
            console.log("User names: " + JSON.stringify(this.usernames));
		},
        selectuser() {
			console.log("Selecting " + this.usertarget);
            if(this.usertarget != "") {
                this.userselected = true;

                // load info for this user
                this.refreshItems();
            }
		},
        adduser() {
            axios.post(
                'http://' + this.$addr + '/api/user',
                {Name: this.newusername},
                axiosConfig)
            .then(response => {
                // add the new user to the users list
                this.users.push(response.data);
                this.makeUserNames();
                this.usercrutch++;
            })
            .catch(error => {
                alert("Failed to add user: " + error);
            });
        },
        additem() {
            axios.post(
                'http://' + this.$addr + '/api/item',
                {
                    Name: this.additemname,
                    ItemType: this.additemtype,
                    Username: this.usertarget,
                    Value: Math.round(100 * parseFloat(this.additemvalue))
                },
                axiosConfig)
            .then(response => {
                // request a new item list
                console.log(response);
                this.refreshItems();
            })
            .catch(error => {
                alert("Failed to add item: " + error);
            });
        },
        deleteitem(id) {
            axios.post(
                'http://' + this.$addr + '/api/itemdelete',
                {
                    Id: id,
                },
                axiosConfig)
            .then(response => {
                // request a new item list
                console.log(response);
                this.refreshItems();
            })
            .catch(error => {
                alert("Failed to delete item: " + error);
            });
        },
	},
}
</script>