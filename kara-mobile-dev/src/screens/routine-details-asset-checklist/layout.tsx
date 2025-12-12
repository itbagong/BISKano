/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-native/no-inline-styles */
/* eslint-disable react-hooks/exhaustive-deps */
import React from 'react';
import {
  ActivityIndicator,
  RefreshControl,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import container, {ContainerContext} from 'components/container';
import {useIsFocused} from '@react-navigation/native';
import s from '@components/styles/index';
import {Colors, Mixins, Typography} from 'utils';
import AccordionItem from 'components/accordion-item';
import {Button, Icon, Input, Layout, Radio} from '@ui-kitten/components';
import {useAppState, useActions} from '@overmind/index';
import {ALERT_TYPE, Toast} from 'react-native-alert-notification';

type Props = {
  route: any;
  navigation: any;
};

const Checklist = (props: Props) => {
  const {navigation, route} = props;
  const isFocused = useIsFocused();
  const ctx = React.useContext(ContainerContext);
  const {routineChecklist} = useAppState();
  const {changeDataRoutineChecklist} = useActions();
  const [data, setData] = React.useState({} as any);
  const [loading, setLoading] = React.useState(false);
  const initData = () => {
    setLoading(true);
    setTimeout(() => {
      const _data =
        routineChecklist?.RoutineChecklistCategories[route.params.index];
      setData({..._data});
      setLoading(false);
    }, 100);
  };
  React.useLayoutEffect(() => {
    navigation.setOptions({
      headerTitle: route.params.headerTitle,
    });
    ctx.setRefreshCallback({
      func: async () => {
        initData();
      },
    });
    return () => {};
  }, [isFocused]);
  React.useEffect(() => {
    initData();
    return () => {};
  }, []);

  const onSave = () => {
    if (data.RoutineChecklistDetails.find((o: any) => o.Status === '')) {
      return Toast.show({
        type: ALERT_TYPE.DANGER,
        title: 'Error',
        textBody:
          'Oops! It seems that some status on the checklist remain unchecked. Please complete all items before proceeding further.',
      });
    }
    const RoutineChecklistCategories = [
      ...routineChecklist?.RoutineChecklistCategories,
    ];
    RoutineChecklistCategories[route.params.index] = data;
    changeDataRoutineChecklist({
      ...routineChecklist,
      RoutineChecklistCategories: RoutineChecklistCategories,
    }).then(() => {
      navigation.goBack();
    });
  };
  return (
    <>
      {isFocused && (
        <View style={s.container}>
          <View style={{flex: 1}}>
            {loading ? (
              <View
                style={{
                  flex: 1,
                  justifyContent: 'center',
                  alignItems: 'center',
                }}>
                <ActivityIndicator color={Colors.PRIMARY.red} />
              </View>
            ) : (
              <ScrollView
                contentInsetAdjustmentBehavior="automatic"
                refreshControl={
                  <RefreshControl
                    refreshing={loading}
                    onRefresh={() => {
                      initData();
                    }}
                  />
                }>
                {data?.RoutineChecklistDetails?.map((item: any, i: number) => (
                  <AccordionItem key={i} title={item.Name} header={(
                    <View style={{flexDirection: 'column', flex: 1}}>
                      <Text style={{
                        ...Typography.textLgSemiBold,
                        color: Colors.BLACK,
                        flex: 1,
                        flexWrap: 'wrap'
                      }}
                      // numberOfLines={1}
                      // ellipsizeMode="tail"
                      >
                        {item.Name}
                      </Text>
                      {/* <Text style={[styles.label]}>Status</Text> */}
                      <Layout
                        style={{
                          flexDirection: 'row',
                          flexWrap: 'wrap',
                          gap: 5,
                        }}
                        level="1">
                        <Radio
                          style={styles.radio}
                          checked={item.Status === 'Normal'}
                          disabled={
                            routineChecklist?.RoutineChecklist
                              ?.IsAlreadyRequest || route?.params?.isReadOnly
                          }
                          onChange={() => {
                            let checklists = [
                              ...data.RoutineChecklistDetails,
                            ];
                            const newItem = {...item, Status: 'Normal'};
                            checklists[i] = newItem;
                            setData({
                              ...data,
                              RoutineChecklistDetails: checklists,
                            });
                          }}>
                          {(evaProps: any) => (
                            <Text
                              style={{
                                ...evaProps.style,
                                ...Typography.textMdPlus,
                              }}>
                              Normal
                            </Text>
                          )}
                        </Radio>
                        <Radio
                          style={styles.radio}
                          checked={item.Status === 'Damaged'}
                          disabled={
                            routineChecklist?.RoutineChecklist
                              ?.IsAlreadyRequest || route?.params?.isReadOnly
                          }
                          onChange={() => {
                            let checklists = [
                              ...data.RoutineChecklistDetails,
                            ];
                            const newItem = {...item, Status: 'Damaged'};
                            checklists[i] = newItem;
                            setData({
                              ...data,
                              RoutineChecklistDetails: checklists,
                            });
                          }}>
                          {(evaProps: any) => (
                            <Text
                              style={{
                                ...evaProps.style,
                                ...Typography.textMdPlus,
                              }}>
                              Damaged
                            </Text>
                          )}
                        </Radio>
                      </Layout>
                    </View>
                  )}>
                    <View>
                      <View style={{marginBottom: Mixins.scaleSize(10)}}>
                        <Text style={[styles.label]}>Code</Text>
                        <Input
                          disabled={
                            routineChecklist?.RoutineChecklist
                              ?.IsAlreadyRequest || route?.params?.isReadOnly
                          }
                          value={item.Code}
                          textStyle={{...Typography.textMdPlus}}
                          placeholder=""
                          onChangeText={nextValue => {
                            let checklists = [...data.RoutineChecklistDetails];
                            const newItem = {...item, Code: nextValue};
                            checklists[i] = newItem;
                            setData({
                              ...data,
                              RoutineChecklistDetails: checklists,
                            });
                          }}
                        />
                      </View>
                      <View style={{marginBottom: Mixins.scaleSize(10)}}>
                        <Text style={[styles.label]}>Note</Text>
                        <Input
                          disabled={
                            routineChecklist?.RoutineChecklist
                              ?.IsAlreadyRequest || route?.params?.isReadOnly
                          }
                          value={item.Note}
                          placeholder="Note"
                          textStyle={{...Typography.textMdPlus}}
                          multiline={true}
                          numberOfLines={4}
                          onChangeText={nextValue => {
                            let checklists = [...data.RoutineChecklistDetails];
                            const newItem = {...item, Note: nextValue};
                            checklists[i] = newItem;
                            setData({
                              ...data,
                              RoutineChecklistDetails: checklists,
                            });
                          }}
                        />
                      </View>
                    </View>
                  </AccordionItem>
                ))}
              </ScrollView>
            )}
          </View>
          {!route?.params?.isReadOnly && (
            <View>
              <Button
                onPress={() => onSave()}
                style={styles.buttonAdd}
                size="large"
                status="primary"
                accessoryLeft={() => (
                  <Icon
                    style={{
                      width: Mixins.scaleSize(22),
                      height: Mixins.scaleSize(22),
                    }}
                    fill="#ffffff"
                    name="save"
                  />
                )}>
                {() => (
                  <Text
                    style={{
                      ...Typography.textLgSemiBold,
                      color: 'white',
                      marginLeft: Mixins.scaleSize(10),
                    }}>
                    Save
                  </Text>
                )}
              </Button>
            </View>
          )}
        </View>
      )}
    </>
  );
};

export default container(Checklist, false);

const styles = StyleSheet.create({
  rowForm: {
    flexDirection: 'row',
  },
  label: {
    ...Typography.textMdPlus,
    color: Colors.BLACK,
    marginBottom: Mixins.scaleSize(5),
  },
  radio: {},
  buttonAdd: {
    borderRadius: Mixins.scaleSize(8),
    alignItems: 'center',
    justifyContent: 'center',
  },
});
