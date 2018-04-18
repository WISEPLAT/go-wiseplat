// Copyright 2015 The go-wiseplat Authors
// This file is part of the go-wiseplat library.
//
// The go-wiseplat library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wiseplat library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wiseplat library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Wiseplat network.
var MainnetBootnodes = []string{
	// Wiseplat Foundation Go Bootnodes
	"enode://14ce82af1caf8368b940ef6212aa7e0ebf6a7bc83ede898bce58cb9cc027663397d25cc4d3bb74d16a1d410e78e7abe1cea185d6aef069767c14da856176f5d8@51.38.35.70:30373",    // WISEPLAT01, France	
	"enode://397604c93dc95294c38f433ed0698b4413b8189d946b1602c6fc59a47f2c35d18d6e347bd690c7c25ac4e07e4a5825b88204910fa3db175b4bf455a16fef85a5@54.37.205.183:30373",  // WISEPLAT02, Germany
	"enode://7cc9c110aa328fd71b18e9fef663ccac580d556290765b6b565a5132d4e319ba3c332c4fe4032c35dab1861c22c6d762260e9bb4f61d23f1b5c325004e4af82c@139.99.96.163:30373",  // WISEPLAT03, Singapore
	
	"enode://4be13b5c7d0bc91376559ada0114b3e8394539e12fb3e98007f1de0548ab3d1982b74230feed35737a9ca8c1b98c7a1b297a61b38c81e23eae4a80876e24385d@52.233.56.177:30373",  // WISEPLAT01, Canada
	"enode://8eb2e97dc4992cf56e8ea78ddf3415a506c9d5159d072a6cb9b128db00ffe631efb07a5b0a43062ac34b99158ac0ad3ea37af78ed139f0f82a6cacc895c478cd@13.89.233.178:30373",  // WISEPLAT02, US
	
	"enode://72a0488e699fce37ad87e25999a90fa11552bc1b7740e71740e98752bc07cb7f569522a9354880c9469fb86491b8359343854144a7186d3ded7d36530dd8a8c1@35.187.122.94:30373",  // WISEPLAT01, Belgium
	"enode://25808971cb173da3261509b39e7b8e5bc1bbca533217ad2158498424b7fbcf5c4624c275914a6592db68cba3537cc11e8925661075af9b58d01b7ecfabfd13dd@35.187.253.180:30373", // WISEPLAT02, Singapore
	
	// Wiseplat Foundation Cpp Bootnodes
	//"enode://979b7fa28feeb35a4741660a16076f1943202cb72b6af70d327f053e248bab9ba81760f39d0701ef1d8f89cc1fbd2cacba0710a12cd5314d5e0c9021aa3637f9@111.111.111.111:30373", // DE
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://96b851f506c32ded38f7ab13ff245d4b409d522230f918ab4b99106a55a021960fae1b72bba865378d86b98bf574bdffd0ebdf46a782147b23bb9a6291d4c6dd@137.74.115.126:30373", // WISEPLAT-TEST-01, France
	//"enode://20c9ad97c081d63397d7b685a412227a40e23c8bdc6688c6f37e97cfbc22d2b4d1db1510d8f61e6a8866ad7f0e17c02b14182d37ea7c3c8b9c2683aeb6b733a1@111.111.111.111:30373", // IE
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{
	"enode://a24ac7c5484ef4ed0c5eb2d36620ba4e4aa13b8c84684e1b4aab0cebea2ae45cb4d375b77eab56516d34bfbd3c1a833fc51296ff084b770b94fb9028c4d25ccf@111.111.111.111:30373", // IE
	//"enode://343149e4feefa15d882d9fe4ac7d88f885bd05ebb735e547f12e12080a9fa07c8014ca6fd7f373123488102fe5e34111f8509cf0b7de3f5b44339c9f25e87cb8@111.111.111.111:30373",  // INFURA
}

// RinkebyV5Bootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network for the experimental RLPx v5 topic-discovery network.
var RinkebyV5Bootnodes = []string{
	"enode://a24ac7c5484ef4ed0c5eb2d36620ba4e4aa13b8c84684e1b4aab0cebea2ae45cb4d375b77eab56516d34bfbd3c1a833fc51296ff084b770b94fb9028c4d25ccf@111.111.111.111:30373?discport=30374", // IE
	//"enode://343149e4feefa15d882d9fe4ac7d88f885bd05ebb735e547f12e12080a9fa07c8014ca6fd7f373123488102fe5e34111f8509cf0b7de3f5b44339c9f25e87cb8@111.111.111.111:30373?discport=30374",  // INFURA
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://0cc5f5ffb5d9098c8b8c62325f3797f56509bff942704687b6530992ac706e2cb946b90a34f1f19548cd3c7baccbcaea354531e5983c7d1bc0dee16ce4b6440b@111.111.111.111:30305",
	//"enode://1c7a64d76c0334b0418c004af2f67c50e36a3be60b5e4790bdac0439d21603469a85fad36f2473c9a80eb043ae60936df905fa28f1ff614c3e5dc34f15dcd2dc@40.118.3.223:30308",
	//"enode://85c85d7143ae8bb96924f2b54f1b3e70d8c4d367af305325d30a61385a432f247d2c75c45c6b4a60335060d072d7f5b35dd1d4c45f76941f62a4f83b6e75daaf@40.118.3.223:30309",
}
