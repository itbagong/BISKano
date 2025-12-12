/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
import {StyleSheet, Text, View} from 'react-native';
import React from 'react';
import s from '@components/styles/index';
import {Avatar, Button, Card} from '@ui-kitten/components/ui';
import {Colors, Mixins, Typography} from 'utils';
import {useActions, useAppState} from '@overmind/index';
import images from 'assets/images';
import FontAwesome from 'react-native-vector-icons/FontAwesome';

type Props = {};

const Layout = (props: Props) => {
  const {} = props;
  const {signOut} = useActions();
  const {userInfo, company} = useAppState();
  return (
    <View style={s.container}>
      <View>
        {/* <Image
          source={{uri: 'https://i.pravatar.cc/400?img=47'}}
          style={styles.avatar}
        /> */}
        <Card
          style={{
            borderRadius: Mixins.scaleSize(10),
            marginBottom: Mixins.scaleSize(10),
          }}>
          {/* <Avatar source={{uri: 'https://i.pravatar.cc/400?img=47'}} /> */}
          <View style={{...s.row, gap: 5, marginBottom: Mixins.scaleSize(10)}}>
            <Avatar size="large" source={images.person} />
            <Text style={styles.username}>{userInfo.DisplayName}</Text>
          </View>
          <View style={styles.infoContainer}>
            <Text style={styles.infoItem}>Company: {company.Name}</Text>
            <Text style={styles.infoItem}>Email: {userInfo.Email}</Text>
          </View>
        </Card>
        <Button
          onPress={() => {
            signOut();
          }}
          style={styles.buttonSignOut}
          size="large"
          status="primary"
          accessoryLeft={() => {
            return (
              <FontAwesome
                name="sign-out"
                size={Mixins.scaleSize(30)}
                color={Colors.WHITE}
              />
            );
          }}>
          {() => (
            <Text style={{...Typography.textLgSemiBold, color: 'white'}}>
              Sign Out
            </Text>
          )}
        </Button>
      </View>
    </View>
  );
};

export default Layout;

const styles = StyleSheet.create({
  username: {
    color: Colors.BLACK,
    ...Typography.textLg,
  },
  infoContainer: {
    backgroundColor: Colors.SHADES.gray[100],
    padding: Mixins.scaleSize(10),
    borderRadius: 10,
  },
  infoItem: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
    marginBottom: Mixins.scaleSize(5),
  },
  buttonSignOut: {
    borderRadius: Mixins.scaleSize(8),
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
  },
});
