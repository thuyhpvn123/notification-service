// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import {NotificationManager} from "../contracts/Event.sol";
import "../contracts/interfaces/INoti.sol";

contract NotificationManagerTest is Test {
  NotificationManager notificationManager;
  address admin;
  address user1;
  address dapp1;
  address dapp2;
  event NotificationSent(
      address indexed dapp,
      address indexed user,
      string title,
      string body,
      uint atTime
  );

  function setUp() public {
      // Set up initial state
      admin = vm.addr(1);
      user1 = vm.addr(2);
      dapp1 = vm.addr(3);
      dapp2 = vm.addr(4);

      // Deploy the NotificationManager contract
      vm.prank(admin);
      notificationManager = new NotificationManager();
      
  }

  function testAddAndRemoveSystemDApp() public {
      vm.startPrank(admin);

      // Add dapp1 as a system DApp
      notificationManager.addSystemDApp(dapp1);
      assertTrue(notificationManager.mSystemDApps(dapp1));

      // Remove dapp1 as a system DApp
      notificationManager.removeSystemDApp(dapp1);
      assertFalse(notificationManager.mSystemDApps(dapp1));
      vm.stopPrank();
  }

  function testRegisterDeviceToken() public {
      string memory encryptedToken = "encrypted-token";

      // Register device token for user1 and dapp1
      vm.prank(user1);
      notificationManager.registerDeviceToken(dapp1, encryptedToken, PlatformEnum.ANDROID);

      // Check that dapp1 is authorized for user1
      assertTrue(notificationManager.mDappUserToPermission(dapp1, user1));
  }

  function testRevokeDAppPermission() public {
      string memory encryptedToken = "encrypted-token";

      // Register device token for user1 and dapp1
      vm.prank(user1);
      notificationManager.registerDeviceToken(dapp1, encryptedToken, PlatformEnum.ANDROID);

      // Revoke permission
      vm.prank(user1);
      notificationManager.revokeDAppPermission(dapp1);

      // Check that dapp1 is no longer authorized for user1
      assertFalse(notificationManager.mDappUserToPermission(dapp1, user1));
  }

  function testAddNotificationFromSystemDApp() public {
      NotiParams memory params = NotiParams("Test Title", "Test Body");

      // Add dapp1 as a system DApp
      vm.prank(admin);
      notificationManager.addSystemDApp(dapp1);

      // Send notification
      vm.prank(dapp1);
      vm.expectEmit(true, true, false, true);
      emit NotificationSent(dapp1, user1, "Test Title", "Test Body", block.timestamp);
      notificationManager.AddNoti(params, user1);

  }
  function testAddNotificationFromDAppPermission() public {
      NotiParams memory params = NotiParams("Test Title", "Test Body");
      string memory encryptedToken = "encrypted-token";

      // Add dapp1 as a system DApp
      vm.prank(user1);
      notificationManager.registerDeviceToken(dapp1, encryptedToken, PlatformEnum.ANDROID);

      // Send notification
      vm.prank(dapp1);
      vm.expectEmit(true, true, false, true);
      emit NotificationSent(dapp1, user1, "Test Title", "Test Body", block.timestamp);
      notificationManager.AddNoti(params, user1);

  }

  function testOnlyAdminCanAddSystemDApp() public {
      // Attempt to add dapp1 as a system DApp by a non-admin
      vm.prank(user1);
      vm.expectRevert("Only admin can perform this action");
      notificationManager.addSystemDApp(dapp1);
  }

  function testOnlyAuthorizedDAppCanSendNotification() public {
      NotiParams memory params = NotiParams("Test Title", "Test Body");

      // Attempt to send a notification from an unauthorized DApp
      vm.prank(dapp2);
      vm.expectRevert("DApp not authorized to send notifications");
      notificationManager.AddNoti(params, user1);
      GetByteCode();
  }
  function GetByteCode()public view {
    address dapp = 0x65e3fcB426Cd39C515880298a6fc15886F1b8a79;
     bytes memory bytesCodeCall = abi.encodeCall(
      notificationManager.addSystemDApp,
          (
            dapp
          )
      );
      console.log("addSystemDApp:");
      console.logBytes(bytesCodeCall);
      console.log(
          "-----------------------------------------------------------------------------"
      );  
      //registerDeviceToken IOS
      string memory encryptedToken = "nruKKK3lvN6MDi++OEAN4x40/UwYhY7TpkQP11jiQPJlCXPbIjHwSh7C4+QKmwxdO2I36/aJD6S4pe4M+xtAySGLe1U28PPhNevOrK3HgG7OkVnnXmlMPK6JJjZDT6U3v6eV4wycImE0SOHGhk4lQMaAtbfxg8r9VpGNc7H5ni2l+CWeZihzPfECNo3PhY5GtmIdv4Cz24/9W3y8qX7KhzKG6ala3Kl5PynWcNqfqMeEg13uMUDZ/Fr5WzWrun5/ni/LIM1Aza3xRVmlC8JoUBbsbuBjK6m5dAefrUDNJBhcIv6Peqjqf8JAaAfGBLnYa0nCj1vknoeFd9qZLeC/7A==:Sqz7q6eJjs64jSRnHJmQa4DFsjphc8qxpM1NJ4zp63bxRVeKVh025NJkiP+ofQFLQo1oFt+QJC9MHFguvK+HwkFOIedcUXGYsGQw3zCluLBx30oBeDnx2TeXWNM=";
      //from token = 8a8ef30b46a858379cbaad6d4a3cbf5defb2ffc8aae2be7cf097377a36a4e682
      bytesCodeCall = abi.encodeCall(
      notificationManager.registerDeviceToken,
          (
            dapp,
            encryptedToken,
            PlatformEnum.IOS
          )
      );
      console.log("registerDeviceToken ios:");
      console.logBytes(bytesCodeCall);
      console.log(
          "-----------------------------------------------------------------------------"
      );  
      //registerDeviceToken Android
      string memory encryptedToken1 ="f1jiERTlT1pSsTuLDxH3J77lQj2boHz0tg7BpAtBZJeUCMrHTfDDIoYIld2wWtKnVNVP1KOOeBLNj4CRI/5o7rOCvaSwYzl+bnL/HxKRCJaB1SJqV21oXyOcHJbx9rcPyNZ4VKSDOZ2N/hCkUOi1PHUWumCIw4ITQD+YsJvG3Jzmq8kU9KSivlhHfv6fyycKMhXeIWA5WZk97c34VyHCBHQMnMOIeFc9Bc49wyMzYm+CVFLo7DEiirhX4gj672Q8Nx8zq9TPNVLE1FR4GA0MjsPYNEjbdnut/x187Vx6BmjFDaRAqe8sZdo0OydIoUUKLPe9bouiX33juoGdnquSDg==:fFiU//ORz8YaIR67xejV1iLMVRKZm0QOLFFGcwiqhSEW9zCUyxRBP2Qv2Tg1TXDSY1mNKwFofziv4FOh9jYeKDltis2prdCE7NU8+HaYa6mv2cLdrAnpi+YB8MoMJVK2qwlF42ScJ0DhtpSuBjO6vtUnCech4633pmBZ79D2UIsm7hZclHGTBzBtnmIdclrZur1sSrv+0wAOmG95lDR7lKtTQ18Qu+WK+5k=";
       //from token = "dBVdZj6-TqmzxR3c6WjNkO:APA91bHJ3Q6B6qu9BX80LmlUIJvW5zUQbiTKxMZBibIz8SfILy8hXqpmjAMmvX2g8xpwhe3V9U6wC4K5bASR6LiI3bBRUZUZmCcnlKoYDmU0KLoxUw2ML3c", 

      bytesCodeCall = abi.encodeCall(
      notificationManager.registerDeviceToken,
          (
            dapp,
            encryptedToken1,
            PlatformEnum.ANDROID
          )
      );
      console.log("registerDeviceToken Android:");
      console.logBytes(bytesCodeCall);
      console.log(
          "-----------------------------------------------------------------------------"
      );  

  }
}
